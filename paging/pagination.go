package paging

import (
	"go.mongodb.org/mongo-driver/bson"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Paging struct {
	Coll       *mongo.Collection
	Condition  map[string]interface{}
	Results    []bson.M
	OptionFind *options.FindOptions
}

type ObjectID struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
}

func NewPaging(coll *mongo.Collection, condition map[string]interface{}, sort bson.M, limit int64) *Paging {
	return &Paging{
		Coll:       coll,
		Condition:  condition,
		OptionFind: setOption(sort, limit)}
}

func NewPagingProjection(coll *mongo.Collection, condition map[string]interface{}, sort bson.M, limit int64, projection bson.M) *Paging {
	return &Paging{
		Coll:       coll,
		Condition:  condition,
		OptionFind: setOptionProject(sort, limit, projection)}
}

func setOption(sort bson.M, limit int64) *options.FindOptions {
	optionsFind := &options.FindOptions{}
	optionsFind.SetSort(sort)
	optionsFind.SetLimit(limit)
	return optionsFind
}

func setOptionProject(sort bson.M, limit int64, projection bson.M) *options.FindOptions {
	optionsFind := &options.FindOptions{}
	optionsFind.SetSort(sort)
	optionsFind.SetLimit(limit)
	optionsFind.SetProjection(projection)
	return optionsFind
}

func (paging *Paging) SetOptions(sort bson.M, limit int64) {
	optionsFind := &options.FindOptions{}
	optionsFind.SetSort(sort)
	optionsFind.SetLimit(limit)
	paging.OptionFind = optionsFind
}

// return error if exist error in decode or error in cursor otherwise nil
func (paging *Paging) Paginage(condition bson.M, t time.Duration) (primitive.ObjectID, error) {
	paging.Results = make([]bson.M, 0, 0)
	lastId := primitive.NewObjectID()
	// default find on _id field
	paging.Condition["_id"] = condition
	ctx, _ := context.WithTimeout(context.Background(), t)
	cur, er := paging.Coll.Find(
		ctx,
		paging.Condition,
		paging.OptionFind)

	if er != nil {
		return primitive.NewObjectID(), er
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println("Error on parsing, ", err.Error())
			continue
		}

		// field forr sort
		if id, ok := result["_id"].(primitive.ObjectID); ok {
			lastId = id
		}
		paging.Results = append(paging.Results, result)
	}
	if err := cur.Err(); err != nil {
		return primitive.NewObjectID(), err
	}

	return lastId, nil
}

// initial maxkey for sort on index field <= current use _id field <= update the lastest or oldest belong to sort
func (paging *Paging) GetMaxKey(projection bson.M, sort bson.M, condition bson.M) (primitive.ObjectID, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	var maxKey ObjectID
	er := paging.Coll.FindOne(
		ctx,
		condition,
		options.FindOne().SetProjection(projection).SetSort(sort)).Decode(&maxKey)

	if er != nil {
		return primitive.NilObjectID, er
	}
	return maxKey.ID, nil
}
