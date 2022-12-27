package core

import (
	"github.com/google/uuid"
	"time"
)

type IdentityRef struct {
	Links             interface{} `json:"_links,omitempty"`
	Descriptor        *string     `json:"descriptor,omitempty"`
	DisplayName       *string     `json:"displayName,omitempty"`
	Url               *string     `json:"url,omitempty"`
	DirectoryAlias    *string     `json:"directoryAlias,omitempty"`
	Id                *string     `json:"id,omitempty"`
	ImageUrl          *string     `json:"imageUrl,omitempty"`
	Inactive          *bool       `json:"inactive,omitempty"`
	IsAadIdentity     *bool       `json:"isAadIdentity,omitempty"`
	IsContainer       *bool       `json:"isContainer,omitempty"`
	IsDeletedInOrigin *bool       `json:"isDeletedInOrigin,omitempty"`
	ProfileUrl        *string     `json:"profileUrl,omitempty"`
	UniqueName        *string     `json:"uniqueName,omitempty"`
}

type ProjectState string

type ProjectVisibility string

type TeamProject struct {
	Abbreviation        *string                       `json:"abbreviation,omitempty"`
	Capabilities        *map[string]map[string]string `json:"capabilities,omitempty"`
	DefaultTeam         *WebApiTeamRef                `json:"defaultTeam,omitempty"`
	DefaultTeamImageUrl *string                       `json:"defaultTeamImageUrl,omitempty"`
	Description         *string                       `json:"description,omitempty"`
	Id                  *uuid.UUID                    `json:"id,omitempty"`
	LastUpdateTime      *Time                         `json:"lastUpdateTime,omitempty"`
	Links               *interface{}                  `json:"_links,omitempty"`
	Name                *string                       `json:"name,omitempty"`
	Revision            *uint64                       `json:"revision,omitempty"`
	State               *ProjectState                 `json:"state,omitempty"`
	Url                 *string                       `json:"url,omitempty"`
	Visibility          *ProjectVisibility            `json:"visibility,omitempty"`
}

type Time struct {
	Time time.Time
}

type WebApiTeamRef struct {
	Id   *uuid.UUID `json:"id,omitempty"`
	Name *string    `json:"name,omitempty"`
	Url  *string    `json:"url,omitempty"`
}
