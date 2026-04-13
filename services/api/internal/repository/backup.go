package repository

import (
	"context"
	"encoding/json"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"github.com/mmtaee/ocserv-dashboard/common/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
)

type BackupRepository struct {
	db *gorm.DB
}

type BackupRepositoryInterface interface {
	OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error
	OcservGroupRestore(ctx context.Context, users *[]models.OcservGroup) (*[]string, *[]string, error)
	OcservUserBackup(ctx context.Context, writer io.Writer) error
	OcservUserRestore(ctx context.Context, users *[]models.OcservUser) (*[]string, *[]string, error)
}

func NewBackupRepository() *BackupRepository {
	return &BackupRepository{
		db: database.GetConnection(),
	}
}

func (b *BackupRepository) OcservGroupBackup(ctx context.Context, writer io.Writer, defaultGroup *models.OcservGroupConfig) error {
	// Start root object
	if _, err := writer.Write([]byte("{")); err != nil {
		return err
	}

	// Write default_group
	if _, err := writer.Write([]byte(`"default_group":`)); err != nil {
		return err
	}

	defaultBytes, err := json.Marshal(defaultGroup)
	if err != nil {
		return err
	}

	if _, err = writer.Write(defaultBytes); err != nil {
		return err
	}

	// Start groups array
	if _, err = writer.Write([]byte(`,"groups":[`)); err != nil {
		return err
	}

	rows, err := b.db.WithContext(ctx).
		Model(&models.OcservGroup{}).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	first := true

	for rows.Next() {
		var group models.OcservGroup

		if err = b.db.ScanRows(rows, &group); err != nil {
			return err
		}

		if !first {
			if _, err = writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		groupBytes, err := json.Marshal(group)
		if err != nil {
			return err
		}

		if _, err = writer.Write(groupBytes); err != nil {
			return err
		}
	}

	// Close array + object
	if _, err := writer.Write([]byte("]}")); err != nil {
		return err
	}

	return nil
}

func (b *BackupRepository) OcservGroupRestore(ctx context.Context, users *[]models.OcservGroup) (*[]string, *[]string, error) {
	names := make([]string, 0, len(*users))
	for _, u := range *users {
		names = append(names, u.Name)
	}

	var dbExisting []string

	err := b.db.WithContext(ctx).
		Model(&models.OcservGroup{}).
		Select("name").
		Where("name IN ?", names).
		Scan(&dbExisting).Error
	if err != nil {
		return nil, nil, err
	}

	existingMap := make(map[string]struct{}, len(dbExisting))
	for _, name := range dbExisting {
		existingMap[name] = struct{}{}
	}

	var toInsert []models.OcservGroup
	var insertedNames []string

	for _, u := range *users {
		if _, found := existingMap[u.Name]; !found {
			toInsert = append(toInsert, u)
			insertedNames = append(insertedNames, u.Name)
		}
	}

	if len(toInsert) > 0 {
		err = b.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toInsert).Error
	}
	return &insertedNames, &dbExisting, err
}

func (b *BackupRepository) OcservUserBackup(ctx context.Context, writer io.Writer) error {
	rows, err := b.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	// Start JSON array
	if _, err = writer.Write([]byte("[")); err != nil {
		return err
	}

	first := true

	for rows.Next() {
		var user models.OcservUser

		if err = b.db.ScanRows(rows, &user); err != nil {
			return err
		}

		if !first {
			if _, err = writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		if _, err = writer.Write(userBytes); err != nil {
			return err
		}
	}

	// Close JSON array
	if _, err = writer.Write([]byte("]")); err != nil {
		return err
	}

	return nil
}

func (b *BackupRepository) OcservUserRestore(ctx context.Context, users *[]models.OcservUser) (*[]string, *[]string, error) {
	usernames := make([]string, 0, len(*users))
	for _, u := range *users {
		usernames = append(usernames, u.Username)
	}

	var dbExisting []string

	err := b.db.WithContext(ctx).
		Model(&models.OcservUser{}).
		Select("username").
		Where("username IN ?", usernames).
		Scan(&dbExisting).Error
	if err != nil {
		return nil, nil, err
	}

	existingMap := make(map[string]struct{}, len(dbExisting))
	for _, name := range dbExisting {
		existingMap[name] = struct{}{}
	}

	var toInsert []models.OcservUser
	var insertedNames []string

	for _, u := range *users {
		if _, found := existingMap[u.Username]; !found {
			toInsert = append(toInsert, u)
			insertedNames = append(insertedNames, u.Username)
		}
	}

	if len(toInsert) > 0 {
		err = b.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toInsert).Error
	}
	return &insertedNames, &dbExisting, err
}
