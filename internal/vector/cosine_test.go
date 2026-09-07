package vector

import (
	"errors"
	"math"
	"testing"
)

func almostEqual(left, right float64) bool {
	const tolerance = 1e-9
	return math.Abs(left-right) < tolerance
}

func TestCosineSimilarityIdenticalVectors(t *testing.T) {
	got, err := CosineSimilarity(
		[]float32{1, 2, 3},
		[]float32{1, 2, 3},
	)
	if err != nil {
		t.Fatalf("CosineSimilarity() returned error: %v", err)
	}

	if !almostEqual(got, 1) {
		t.Fatalf("similarity = %f, want 1", got)
	}
}

func TestCosineSimilarityOrthogonalVectors(t *testing.T) {
	got, err := CosineSimilarity(
		[]float32{1, 0},
		[]float32{0, 1},
	)
	if err != nil {
		t.Fatalf("CosineSimilarity() returned error: %v", err)
	}

	if !almostEqual(got, 0) {
		t.Fatalf("similarity = %f, want 0", got)
	}
}

func TestCosineSimilarityOppositeVectors(t *testing.T) {
	got, err := CosineSimilarity(
		[]float32{1, 0},
		[]float32{-1, 0},
	)
	if err != nil {
		t.Fatalf("CosineSimilarity() returned error: %v", err)
	}

	if !almostEqual(got, -1) {
		t.Fatalf("similarity = %f, want -1", got)
	}
}

func TestCosineSimilarityRejectsDifferentDimensions(t *testing.T) {
	_, err := CosineSimilarity(
		[]float32{1, 2},
		[]float32{1, 2, 3},
	)
	if !errors.Is(err, ErrDimensionMismatch) {
		t.Fatalf("expected ErrDimensionMismatch, got %v", err)
	}
}

func TestCosineSimilarityRejectsZeroVector(t *testing.T) {
	_, err := CosineSimilarity(
		[]float32{0, 0},
		[]float32{1, 2},
	)
	if !errors.Is(err, ErrZeroVector) {
		t.Fatalf("expected ErrZeroVector, got %v", err)
	}
}
