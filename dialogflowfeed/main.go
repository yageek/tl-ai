package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gophersch/tlgo"
	"google.golang.org/api/option"

	"cloud.google.com/go/dialogflow/apiv2"
	"google.golang.org/api/iterator"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

const (
	LineEntityName = "line-name"
	StopEntityName = "stop-name"
)

var (
	PROJECT_ID    string
	JSON_KEY_PATH string
)

func filter(s string) string {

	invalidCharacters := "(),"

	out := bytes.NewBufferString("")

	for _, c := range s {
		if strings.IndexRune(invalidCharacters, c) < 0 {
			out.WriteRune(c)
		}
	}

	return strings.Replace(out.String(), "VD", "", -1)
}
func main() {

	// Get configuration vars
	PROJECT_ID = os.Getenv("PROJECT_ID")
	JSON_KEY_PATH = os.Getenv("JSON_KEY_PATH")

	if PROJECT_ID == "" || JSON_KEY_PATH == "" {
		log.Fatalf("Missing PROJECT_ID and JSON_KEY_PATH env variable to start.")
	}
	// Retrieve the TL information
	tlClient := tlgo.NewClient()

	// Retrieve all the elements from the TL

	ctx := context.Background()
	bytes, err := ioutil.ReadFile(JSON_KEY_PATH)
	if err != nil {
		log.Fatalf("Can not read the provided JSON key file: %v", err)
	}

	c, err := dialogflow.NewEntityTypesClient(ctx, option.WithCredentialsJSON(bytes))
	if err != nil {
		log.Fatalf("Impossible to create a dialogflowclient: %v", err)

	}

	// Check if previous entity exists
	entities, err := existingEntitiesName(ctx, c)
	if err != nil {
		log.Fatalf("Impossible to create a get entities: %v", err)
	}

	if entity, hasLine := entities[LineEntityName]; hasLine {
		log.Printf("Previous line entity defined, deleting them...")
		if err := deleteEntity(ctx, c, entity); err != nil {
			log.Fatalf("Impossible to delete stop entity: %v", err)
		}
	}

	// Populate line
	modelLines, err := tlClient.ListLines()
	if err != nil {
		log.Fatalf("Impossible to get lines: %s", err)
	}

	modelStops, err := tlClient.ListStops()
	if err != nil {
		log.Fatalf("Can not fetch stop lists: %s", err)
	}

	lineEntities := make([]*dialogflowpb.EntityType_Entity, len(modelLines))
	for i, line := range modelLines {

		lineEntities[i] = &dialogflowpb.EntityType_Entity{
			Value:    line.ShortName,
			Synonyms: []string{filter(line.ShortName)},
		}
	}

	lineEntity := &dialogflowpb.EntityType{
		DisplayName: LineEntityName,
		Entities:    lineEntities,
		Kind:        dialogflowpb.EntityType_KIND_MAP,
	}

	lineEntity, err = createEmptyEntity(ctx, c, lineEntity)
	if err != nil {
		log.Fatalf("Impossible to create the line entity: %v", err)
	}

	// Populate stop
	if stopEntity, hasStop := entities[StopEntityName]; hasStop {
		log.Printf("Previous stop entity defined, deleting them...")
		if err := deleteEntity(ctx, c, stopEntity); err != nil {
			log.Fatalf("Impossible to delete stop entity: %v", err)
		}
	}

	stopEntities := make([]*dialogflowpb.EntityType_Entity, len(modelStops))
	for i, stop := range modelStops {

		stopEntities[i] = &dialogflowpb.EntityType_Entity{
			Value:    stop.Name,
			Synonyms: []string{filter(stop.Name)},
		}
	}

	stopEntity := &dialogflowpb.EntityType{
		DisplayName: StopEntityName,
		Entities:    stopEntities,
		Kind:        dialogflowpb.EntityType_KIND_MAP,
	}

	stopEntity, err = createEmptyEntity(ctx, c, stopEntity)
	if err != nil {
		log.Fatalf("Impossible to create the stop entity: %v", err)
	}

}

func createEmptyEntity(ctx context.Context, c *dialogflow.EntityTypesClient, entity *dialogflowpb.EntityType) (*dialogflowpb.EntityType, error) {
	req := &dialogflowpb.CreateEntityTypeRequest{
		Parent:     fmt.Sprintf("projects/%s/agent", PROJECT_ID),
		EntityType: entity,
	}

	return c.CreateEntityType(ctx, req)
}
func deleteEntity(ctx context.Context, c *dialogflow.EntityTypesClient, entity *dialogflowpb.EntityType) error {
	req := &dialogflowpb.DeleteEntityTypeRequest{
		Name: entity.Name,
	}
	return c.DeleteEntityType(ctx, req)
}
func existingEntitiesName(ctx context.Context, c *dialogflow.EntityTypesClient) (map[string]*dialogflowpb.EntityType, error) {

	entities := make(map[string]*dialogflowpb.EntityType, 1)

	req := &dialogflowpb.ListEntityTypesRequest{
		LanguageCode: "fr",
		Parent:       fmt.Sprintf("projects/%s/agent", PROJECT_ID),
	}
	it := c.ListEntityTypes(ctx, req)

	for {
		resp, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return map[string]*dialogflowpb.EntityType{}, err
		}
		entities[resp.DisplayName] = resp
	}

	return entities, nil
}
