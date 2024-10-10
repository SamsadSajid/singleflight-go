package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "net/http/pprof"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/singleflight"
)

type Customer struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type App struct {
	db     *dynamodb.Client
	cache  *redis.Client
	sf     *singleflight.Group
	apmSvc prometheus.Counter
}

func init() {
	// Create DynamoDB table if it doesn't exist
	db := connectDynamoDB()
	_, err := db.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(20000),
			WriteCapacityUnits: aws.Int64(2000),
		},
		TableName: aws.String("Customers"),
	})

	if err != nil {
		fmt.Println("Table may already exist:", err)
	}
	log.Println("Table created")

	// Insert some sample data
	customers := []Customer{
		{ID: uuid.New().String(), Status: "active"},
		{ID: uuid.New().String(), Status: "not active"},
	}

	for _, c := range customers {
		_, err := db.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String("Customers"),
			Item: map[string]types.AttributeValue{
				"ID":     &types.AttributeValueMemberS{Value: c.ID},
				"Status": &types.AttributeValueMemberS{Value: c.Status},
			},
		})
		if err != nil {
			fmt.Println("Error inserting item:", err)
		} else {
			fmt.Printf("Inserted customer: %s\n", c.ID)
		}
	}
	log.Println("Customers table populated with seed data")
}

func main() {
	app := &App{
		db:    connectDynamoDB(),
		cache: connectRedis(),
		sf:    &singleflight.Group{},
		apmSvc: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_hydration_counter",
			Help: "The total number of cache hydrate events",
		}),
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/customer/{id}", app.getCustomerStatus)
	http.HandleFunc("/customer/{id}/singleflight", app.getCustomerStatusSingleflight)

	// Benchmark using ab tool
	// Run these commands in the terminal:
	// ab -n 100000 -c 100 http://localhost:8080/customer/some-id
	// ab -n 10000 -c 100 http://localhost:8080/customer/some-id/singleflight

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func connectDynamoDB() *dynamodb.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:8000", // Adjust this URL if your local DynamoDB is on a different port
			SigningRegion: "eu-west-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("local"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")),
	)
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func connectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		PoolSize:     1000,            // Set the maximum number of connections in the pool to handle 1000 concurrent connections
		MinIdleConns: 100,             // Minimum number of idle connections to maintain in the pool
		MaxConnAge:   time.Hour,       // Maximum age of a connection in the pool
		IdleTimeout:  5 * time.Minute, // How long a connection can be idle before being closed
	})
}

func (app *App) getCustomerStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	customer, err := app.getCustomer(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(customer)
}

func (app *App) getCustomerStatusSingleflight(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	customer, err := app.getCustomerSingleflight(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(customer)
}

func (app *App) getCustomer(ctx context.Context, id string) (*Customer, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled or deadline exceeded: %w", ctx.Err())
	default:
		customer, err := app.getCustomerFromCache(ctx, id)
		if err == nil {
			log.Println("Cache hit")
			return customer, nil
		}

		// Cache miss, fetch from DynamoDB
		customer, err = app.getCustomerFromDB(ctx, id)
		if err != nil {
			return nil, err
		}

		// Cache the result
		if err := app.cacheCustomer(ctx, id, customer); err != nil {
			// Log the error but don't fail the request
			log.Printf("Failed to cache customer: %v", err)
		}

		return customer, nil
	}
}

func (app *App) getCustomerFromCache(ctx context.Context, id string) (*Customer, error) {
	val, err := app.cache.Get(ctx, id).Result()
	if err != nil {
		return nil, err
	}

	var customer Customer
	if err := json.Unmarshal([]byte(val), &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

func (app *App) getCustomerFromDB(ctx context.Context, id string) (*Customer, error) {
	result, err := app.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("Customers"),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	return &Customer{
		ID:     result.Item["ID"].(*types.AttributeValueMemberS).Value,
		Status: result.Item["Status"].(*types.AttributeValueMemberS).Value,
	}, nil
}

func (app *App) cacheCustomer(ctx context.Context, id string, customer *Customer) error {
	customerJSON, err := json.Marshal(customer)
	if err != nil {
		log.Printf("Failed to marshal customer: %v", err)
		return err
	}
	_, err = app.cache.Set(ctx, id, string(customerJSON), time.Millisecond).Result()
	if err == nil {
		app.apmSvc.Inc()
	}
	return err
}

func (app *App) getCustomerSingleflight(ctx context.Context, id string) (*Customer, error) {
	v, err, _ := app.sf.Do(id, func() (interface{}, error) {
		return app.getCustomer(ctx, id)
	})

	if err != nil {
		return nil, err
	}

	return v.(*Customer), nil
}
