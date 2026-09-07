package vector

import (
	"errors"
	"math"
)

var (
	ErrDimensionMismatch = errors.New(
		"vectors must have the same dimension",
	)
	ErrZeroVector = errors.New(
		"cosine similarity is undefined for a zero vector",
	)
)

func CosineSimilarity(
	left []float32,
	right []float32,
) (float64, error) {
	if len(left) != len(right) {
		return 0, ErrDimensionMismatch
	}

	if len(left) == 0 || len(right) == 0 {
		return 0, ErrZeroVector
	}

	var dotProduct float64
	var leftMagnitudeSquared float64
	var rightMagnitudeSquared float64

	for index := range left {
		leftValue := float64(left[index])
		rightValue := float64(right[index])

		dotProduct += leftValue * rightValue
		leftMagnitudeSquared += leftValue * leftValue
		rightMagnitudeSquared += rightValue * rightValue
	}

	if leftMagnitudeSquared == 0 || rightMagnitudeSquared == 0 {
		return 0, ErrZeroVector
	}

	similarity := dotProduct / (math.Sqrt(leftMagnitudeSquared) *
		math.Sqrt(rightMagnitudeSquared))

	return similarity, nil

}
