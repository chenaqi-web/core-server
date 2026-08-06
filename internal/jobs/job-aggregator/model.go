package jobaggregator

import "fmt"

type ObjectCountDelta struct {
	InteractionType string
	ObjectType      string
	ObjectID        uint64
	Delta           int64
}

func getObjectCountDeltaKey(interactionType, objectType string, objectID uint64) string {
	return fmt.Sprintf("%s:%s:%d", interactionType, objectType, objectID)
}
