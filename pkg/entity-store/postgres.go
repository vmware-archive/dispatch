///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package entitystore

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type postgresEntityStore struct {
	db *sqlx.DB
}

type dbEntity struct {
	Key            string         `db:"key"`
	ID             string         `db:"id"`
	Name           string         `db:"name"`
	Type           string         `db:"type"`
	OrganizationID string         `db:"organization_id"`
	CreatedTime    time.Time      `db:"created_time"`
	ModifiedTime   time.Time      `db:"modified_time"`
	Revision       uint64         `db:"revision"`
	Version        uint64         `db:"version"`
	Status         string         `db:"status"`
	Delete         bool           `db:"delete"`
	Reason         Reason         `db:"reason"`
	Spec           Spec           `db:"spec"`
	Tags           Tags           `db:"tags"`
	Value          types.JSONText `db:"value"`
}

// the value and scan methods listed in the following is used to
// serialize and unserialize the type Reason,  Spec and Tags into and from postgres JSONB columns
func value(v interface{}) (driver.Value, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrapf(err, "error marshalling type=%s", reflect.TypeOf(v))
	}
	return j, nil
}

func scan(v interface{}, src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.Errorf("error scanning type=%s: type assertion .([]byte) failed", reflect.TypeOf(v))
	}
	err := json.Unmarshal(source, v)
	if err != nil {
		return errors.Wrapf(err, "error scanning type=%s unmarshal failed", reflect.TypeOf(v))
	}
	return nil
}

// Value implements interface driver.Valuer
func (v Reason) Value() (driver.Value, error) {
	return value(v)
}

// Scan implements interface sql.Scanner
func (v *Reason) Scan(src interface{}) error {
	return scan(v, src)
}

// Value implements interface driver.Valuer
func (v Spec) Value() (driver.Value, error) {
	return value(v)
}

// Scan implements interface sql.Scanner
func (v *Spec) Scan(src interface{}) error {
	return scan(v, src)
}

// Value implements interface driver.Valuer
func (v Tags) Value() (driver.Value, error) {
	return value(v)
}

// Scan implements interface sql.Scanner
func (v *Tags) Scan(src interface{}) error {
	return scan(v, src)
}

func (p *postgresEntityStore) createTable() error {

	sql := `
		CREATE TABLE IF NOT EXISTS entity (
		key 			TEXT PRIMARY KEY,
		id 				TEXT,
		name 			TEXT,
		type			TEXT,
		organization_id TEXT,
		created_time 	TIME,
		modified_time 	TIME,
		revision 		BIGINT,
		version 		BIGINT,
		status 			TEXT,
		delete 			TEXT,
		spec 			JSONB,
		reason 			JSONB,
		tags			JSONB,
		value 			JSONB
	)`
	_, err := p.db.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "fail to create the entity table")
	}
	return nil
}

func (p *postgresEntityStore) dropTable() error {
	sql := `
	DROP TABLE IF EXISTS entity`
	_, err := p.db.Exec(sql)
	if err != nil {
		log.Debug(err)
		return errors.Wrap(err, "fail to drop the entity table")
	}
	return nil
}

func getHostAndPort(addr string) (string, string, error) {
	res := strings.Split(addr, ":")
	if len(res) != 2 {
		return "", "", fmt.Errorf("Invalid Address")
	}
	return res[0], res[1], nil
}

// newPostgres creates a postgres entity store
func newPostgres(config BackendConfig) (EntityStore, error) {

	opts := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", config.Username, config.Password, config.Address, config.Bucket)
	log.Debugf("postgresql database options: %s", opts)
	db, err := sqlx.Connect("postgres", opts)
	if err != nil {
		log.Debugf("error connecting to postgresql DB")
		return nil, errors.Wrap(err, "Unable to connect to the postgres db server")
	}
	store := &postgresEntityStore{db: db}

	// create tables if not exists
	err = store.createTable()
	if err != nil {
		return nil, err
	}
	return store, nil
}

func dbToEntity(row dbEntity, entity Entity) error {

	err := row.Value.Unmarshal(entity)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling from dbEntity to entity")
	}
	entity.setID(row.ID)
	entity.setName(row.Name)
	entity.setOrganizationID(row.OrganizationID)
	entity.setCreatedTime(row.CreatedTime)
	entity.setModifiedTime(row.ModifiedTime)
	entity.setRevision(row.Revision)
	entity.setVersion(row.Version)
	entity.SetStatus(Status(row.Status))
	entity.SetDelete(row.Delete)
	entity.SetSpec(row.Spec)
	entity.SetReason(row.Reason)
	entity.SetTags(row.Tags)
	return nil
}

func entityToDbEntity(e Entity) (*dbEntity, error) {

	row := &dbEntity{
		Key:            getKey(e),
		ID:             e.GetID(),
		Name:           e.GetName(),
		Type:           string(getDataType(e)),
		OrganizationID: e.GetOrganizationID(),
		CreatedTime:    e.GetCreateTime(),
		ModifiedTime:   e.GetModifiedTime(),
		Revision:       e.GetRevision(),
		Version:        e.GetVersion(),
		Status:         string(e.GetStatus()),
		Spec:           e.GetSpec(),
		Reason:         e.GetReason(),
		Tags:           e.GetTags(),
		Delete:         e.GetDelete(),
		Value:          types.JSONText{},
	}

	value, err := json.Marshal(e)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling entity")
	}
	err = row.Value.Scan(value)
	if err != nil {
		log.Debugf("error scanning entity: %s", err)
		return nil, errors.Wrap(err, "error scanning entity")
	}
	return row, nil
}

func (p *postgresEntityStore) Add(entity Entity) (id string, err error) {

	err = precondition(entity)
	if err != nil {
		return "", errors.Wrap(err, "Precondition failed")
	}
	id = uuid.NewV4().String()
	entity.setID(id)
	now := time.Now()
	entity.setCreatedTime(now)
	entity.setModifiedTime(now)
	row, err := entityToDbEntity(entity)
	if err != nil {
		return "", err
	}
	sql := `INSERT INTO entity
		(key, id, name, type, organization_id, created_time, modified_time, revision, version,
		spec, status, reason, tags, delete, value)
	VALUES
		(:key, :id, :name, :type, :organization_id, :created_time, :modified_time, :revision, :version,
		:spec, :status, :reason, :tags, :delete, :value)`
	_, err = p.db.NamedExec(sql, row)
	if err != nil {
		return "", errors.Wrap(err, "error adding entity into db")
	}
	return id, nil
}

// Update updates existing entities to the store
func (p *postgresEntityStore) Update(lastRevision uint64, entity Entity) (revision int64, err error) {

	entity.setModifiedTime(time.Now())
	entity.setRevision(lastRevision)
	sql := `
	UPDATE entity
	SET
		id = :id, name = :name, organization_id = :organization_id, created_time = :created_time,
		modified_time = :modified_time, revision = revision + 1, version = :version,
		spec = :spec, status = :status, reason = :reason , tags = :tags, delete = :delete, value = :value
	WHERE
		key = :key AND
		revision = :revision
	`
	row, err := entityToDbEntity(entity)
	if err != nil {
		return 0, err
	}

	result, err := p.db.NamedExec(sql, row)
	if err != nil {
		return 0, errors.Wrap(err, "error updating entity")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error updating entity")
	}
	if rowsAffected != 1 {
		return 0, errors.Errorf("error updating entity: no such entity or there's intermidate update")
	}
	entity.setRevision(lastRevision + 1)
	return int64(entity.GetRevision()), nil
}

// Get gets a single entity by key from the store
func (p *postgresEntityStore) Get(organizationID string, name string, opts Options, entity Entity) error {

	key := buildKey(organizationID, getDataType(entity), name)
	if opts.Filter == nil {
		opts.Filter = FilterEverything()
	}
	opts.Filter.Add(FilterStat{
		Scope:   FilterScopeField,
		Subject: "Key",
		Verb:    FilterVerbEqual,
		Object:  key,
	})

	sql, args, err := makeListQuery(organizationID, opts.Filter, reflect.TypeOf(entity).Elem())
	if err != nil {
		return errors.Wrap(err, "error makeListQuery")
	}
	sql = p.db.Rebind(sql)
	rows, err := p.db.Queryx(sql, args...)
	if err != nil {
		return errors.Wrap(err, "error getting: ")
	}

	if rows.Next() == false {
		return errors.New("error getting: no such entity")
	}
	row := dbEntity{}
	err = rows.StructScan(&row)
	if err != nil {
		return errors.Wrap(err, "error getting entity from db")
	}
	if rows.Next() != false {
		return errors.New("error getting: get more than one entity")
	}
	return dbToEntity(row, entity)
}

func makeListQuery(organizationID string, filter Filter, entityType reflect.Type) (sql string, args []interface{}, err error) {

	sql = ""
	argsMap := map[string]interface{}{
		"organization_id": organizationID,
		"type":            dataType(entityType.Name()),
	}
	where := []string{
		"organization_id = :organization_id",
		"type = :type",
	}
	if filter != nil {
		for _, fs := range filter.FilterStats() {
			column := ""
			object := ""
			switch fs.Scope {
			case FilterScopeField:
				field, ok := reflect.TypeOf(dbEntity{}).FieldByName(fs.Subject)
				if !ok {
					err = errors.Errorf("error listing: no such field: %s", fs.Subject)
					return
				}
				// find the column name by struct tag
				column = field.Tag.Get("db")
				object = column
			case FilterScopeTag:
				object = fmt.Sprintf("tag_%s", fs.Subject)
				column = fmt.Sprintf("tags->>'%s'", fs.Subject)
			case FilterScopeExtra:
				field, ok := entityType.FieldByName(fs.Subject)
				if !ok {
					err = errors.Errorf("error listing: no such extra field: %s", fs.Subject)
					return
				}
				// remove the "omitempty"
				object = strings.Split(field.Tag.Get("json"), ",")[0]
				// the value is inside the JSONB field 'value'
				column = fmt.Sprintf("value->>'%s'", object)
			}
			argsMap[object] = fs.Object

			switch fs.Verb {
			case FilterVerbEqual:
				where = append(where, fmt.Sprintf("%s = :%s", column, object))
			case FilterVerbIn:
				where = append(where, fmt.Sprintf("%s IN (:%s)", column, object))
			case FilterVerbBefore:
				where = append(where, fmt.Sprintf("%s < :%s", column, object))
			case FilterVerbAfter:
				where = append(where, fmt.Sprintf("%s > :%s", column, object))
			default:
				err = errors.Errorf("error listing: invalid filter")
				return
			}
		}
	}
	sql = fmt.Sprintf("SELECT * FROM entity WHERE %s", strings.Join(where, " AND "))
	sql, args, err = sqlx.Named(sql, argsMap)
	if err != nil {
		err = errors.Wrap(err, "error making sql query: sqlx.Named")
		return
	}
	sql, args, err = sqlx.In(sql, args...)
	if err != nil {
		err = errors.Wrap(err, "error making sql query: sqlx.In")
		return
	}
	return
}

// List fetches a list of entities of a single data type satisfying the filter.
// entities is a placeholder for results and must be a pointer to an empty slice of the desired entity type.
func (p *postgresEntityStore) List(organizationID string, opts Options, entities interface{}) error {

	rv := reflect.ValueOf(entities)
	if entities == nil || rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return errors.New("need a non-nil entity slice pointer")
	}

	entityPtrType := rv.Elem().Type().Elem()
	if !entityPtrType.Implements(reflect.TypeOf((*Entity)(nil)).Elem()) {
		return errors.New("non-entity element type: maybe use pointers")
	}

	sql, args, err := makeListQuery(organizationID, opts.Filter, entityPtrType.Elem())
	if err != nil {
		return errors.Wrap(err, "error makeListQuery")
	}

	sql = p.db.Rebind(sql)
	rows, err := p.db.Queryx(sql, args...)
	if err != nil {
		return errors.Wrap(err, "error listing entity from db")
	}

	slice := reflect.MakeSlice(rv.Elem().Type(), 0, 0)
	for rows.Next() {
		row := dbEntity{}
		err = rows.StructScan(&row)
		if err != nil {
			return errors.Wrap(err, "error listing entity from db")
		}

		entityPtr := reflect.New(entityPtrType.Elem())
		entity := entityPtr.Interface().(Entity)

		dbToEntity(row, entity)
		slice = reflect.Append(slice, entityPtr)
	}
	rv.Elem().Set(slice)
	return nil
}

// Delete deletes a single entity from the store.
func (p *postgresEntityStore) Delete(organizationID string, name string, entity Entity) error {

	key := buildKey(organizationID, getDataType(entity), name)

	sql := `DELETE FROM entity WHERE key = $1`
	result, err := p.db.Exec(sql, key)
	if err != nil {
		return errors.Wrap(err, "error deleting an entity")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error deleting an entity")
	}
	if rowsAffected == 0 {
		return errors.New("error deleting: no such entity")
	}
	if rowsAffected > 1 {
		return errors.New("error deleting: deleted mutiple entities")
	}
	return nil
}

// UpdateWithError is used by entity handlers to save changes and/or error status
// e.g. `defer func() { h.store.UpdateWithError(e, err) }()`
func (p *postgresEntityStore) UpdateWithError(e Entity, err error) {

	if err != nil {
		e.SetStatus(StatusERROR)
		e.SetReason([]string{err.Error()})
	}
	if _, err2 := p.Update(e.GetRevision(), e); err2 != nil {
		log.Error(err2)
	}
	return
}
