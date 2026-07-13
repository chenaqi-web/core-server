package jobaggregator

import "fmt"

type ObjectCountDelta struct {
	InteractionType string
	ObjectType      string
	ObjectID        string
	Delta           int64
}

func getObjectCountDeltaKey(interactionType, objectType, objectID string) string {
	return fmt.Sprintf("%s:%s:%s", interactionType, objectType, objectID)
}
