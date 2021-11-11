// Package dirstore for storage of directory information
// This is an interface to support different backend implementations
package dirstore

// Interface to the directory JSON object store
// Simple CRUD interface with JSONPATH support
type IDirStore interface {
	// Close the store
	Close()
	// Get a document by its ID
	// Returns an error if it doesn't exist
	Get(id string) (interface{}, error)

	// Get a list of documents
	//  offset to start
	//  limit is the maximum nr of documents to return
	//	filter is a function to filter things
	List(offset int, limit int, filter func(thingID string) bool) []interface{}

	// Open the store
	// Returns error if it can't be opened or already open
	Open() error

	// Patch part of a document
	// Returns an error if it doesn't exist
	Patch(id string, doc map[string]interface{}) error

	// Query for documents using JSONPATH
	//  offset to return the results
	//  maximum nr of documents to return
	//	filter is a function to filter things
	// Returns list of documents by their ID, or error if jsonPath is invalid
	Query(jsonPath string, offset int, limit int, filter func(thingID string) bool) ([]interface{}, error)

	// Remove a document
	// Succeeds if the document doesn't exist
	Remove(id string)

	// Replace a document
	// The document does not have to exist
	Replace(id string, document map[string]interface{}) error
}
