package mongodb

import (
	"Report-Storage/internal/storage"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateReport полностью заменяет заявку на переданную по ее уникальному
// номеру. Аргумент rep должен содержать все поля заявки с корректными
// значениями. Возвращает заявку ДО ее изменения, либо ошибку. Если документ
// с указанным номером не найден, то вернет ошибку ErrReportNotFound.
func (s *Storage) UpdateReport(ctx context.Context, rep storage.Report) (storage.Report, error) {
	const operation = "storage.mongodb.UpdateReport"

	// origin будет содержать заявку до ее изменения.
	var origin storage.Report

	if rep.Number < 1 {
		return origin, fmt.Errorf("%s: %w", operation, storage.ErrIncorrectNum)
	}
	if _, err := primitive.ObjectIDFromHex(rep.ID.Hex()); err != nil {
		return origin, fmt.Errorf("%s: %w", operation, storage.ErrIncorrectID)
	}
	if !checkStatus(rep.Status) {
		return origin, fmt.Errorf("%s: %w", operation, storage.ErrIncorrectStatus)
	}

	// Меняем местами широту и долготу у обновленной заявки.
	rep.Geo.Coordinates[0], rep.Geo.Coordinates[1] = rep.Geo.Coordinates[1], rep.Geo.Coordinates[0]

	collection := s.db.Database(dbName).Collection(colReport)
	filter := bson.D{{Key: "number", Value: rep.Number}}

	err := collection.FindOneAndReplace(ctx, filter, rep).Decode(&origin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return origin, fmt.Errorf("%s: %w", operation, storage.ErrReportNotFound)
		}
		return origin, fmt.Errorf("%s: %w", operation, err)
	}

	// Меняем местами широту и долготу у заявки до изменения.
	origin.Geo.Coordinates[0], origin.Geo.Coordinates[1] = origin.Geo.Coordinates[1], origin.Geo.Coordinates[0]

	return origin, nil
}
