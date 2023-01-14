package workitems

import "github.com/google/uuid"

type WorkItemClassificationNode struct {
	Attributes    *map[string]interface{}       `json:"attributes,omitempty"`
	Children      *[]WorkItemClassificationNode `json:"children,omitempty"`
	HasChildren   *bool                         `json:"hasChildren,omitempty"`
	Id            *int                          `json:"id,omitempty"`
	Identifier    *uuid.UUID                    `json:"identifier,omitempty"`
	Links         interface{}                   `json:"_links,omitempty"`
	Name          *string                       `json:"name,omitempty"`
	Path          *string                       `json:"path,omitempty"`
	StructureType *string                       `json:"structureType,omitempty"`
	Url           *string                       `json:"url,omitempty"`
}
