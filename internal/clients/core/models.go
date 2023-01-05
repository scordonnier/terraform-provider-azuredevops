package core

import (
	"github.com/google/uuid"
	"time"
)

const (
	CapabilitiesProcessTemplate       = "processTemplate"
	CapabilitiesProcessTemplateTypeId = "templateTypeId"
	CapabilitiesVersionControl        = "versioncontrol"
	CapabilitiesVersionControlType    = "sourceControlType"
)

type Identity struct {
	CustomDisplayName   *string      `json:"customDisplayName,omitempty"`
	Descriptor          *string      `json:"descriptor,omitempty"`
	Id                  *uuid.UUID   `json:"id,omitempty"`
	IsActive            *bool        `json:"isActive,omitempty"`
	IsContainer         *bool        `json:"isContainer,omitempty"`
	MasterId            *uuid.UUID   `json:"masterId,omitempty"`
	MemberIds           *[]uuid.UUID `json:"memberIds,omitempty"`
	MemberOf            *[]string    `json:"memberOf,omitempty"`
	Members             *[]string    `json:"members,omitempty"`
	MetaTypeId          *int         `json:"metaTypeId,omitempty"`
	Properties          interface{}  `json:"properties,omitempty"`
	ProviderDisplayName *string      `json:"providerDisplayName,omitempty"`
	ResourceVersion     *int         `json:"resourceVersion,omitempty"`
	SocialDescriptor    *string      `json:"socialDescriptor,omitempty"`
	SubjectDescriptor   *string      `json:"subjectDescriptor,omitempty"`
	UniqueUserId        *int         `json:"uniqueUserId,omitempty"`
}

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

type Operation struct {
	DetailedMessage *string                   `json:"detailedMessage,omitempty"`
	Id              *uuid.UUID                `json:"id,omitempty"`
	Links           interface{}               `json:"_links,omitempty"`
	PluginId        *uuid.UUID                `json:"pluginId,omitempty"`
	ResultMessage   *string                   `json:"resultMessage,omitempty"`
	ResultUrl       *OperationResultReference `json:"resultUrl,omitempty"`
	Status          *string                   `json:"status,omitempty"`
	Url             *string                   `json:"url,omitempty"`
}

type OperationReference struct {
	Id       *uuid.UUID `json:"id,omitempty"`
	PluginId *uuid.UUID `json:"pluginId,omitempty"`
	Status   *string    `json:"status,omitempty"`
	Url      *string    `json:"url,omitempty"`
}

type OperationResultReference struct {
	ResultUrl *string `json:"resultUrl,omitempty"`
}

type Process struct {
	Description *string     `json:"description,omitempty"`
	Id          *uuid.UUID  `json:"id,omitempty"`
	IsDefault   *bool       `json:"isDefault,omitempty"`
	Links       interface{} `json:"_links,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Type        *string     `json:"type,omitempty"`
	Url         *string     `json:"url,omitempty"`
}

type ProcessCollection struct {
	Count *int       `json:"count"`
	Value *[]Process `json:"value"`
}

type ProjectReference struct {
	Id   *uuid.UUID `json:"id,omitempty"`
	Name *string    `json:"name,omitempty"`
}

type ProjectState string

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
	Visibility          *string                       `json:"visibility,omitempty"`
}

type TeamProjectReference struct {
	Abbreviation        *string       `json:"abbreviation,omitempty"`
	DefaultTeamImageUrl *string       `json:"defaultTeamImageUrl,omitempty"`
	Description         *string       `json:"description,omitempty"`
	Id                  *uuid.UUID    `json:"id,omitempty"`
	LastUpdateTime      *Time         `json:"lastUpdateTime,omitempty"`
	Name                *string       `json:"name,omitempty"`
	Revision            *uint64       `json:"revision,omitempty"`
	State               *ProjectState `json:"state,omitempty"`
	Url                 *string       `json:"url,omitempty"`
	Visibility          *string       `json:"visibility,omitempty"`
}

type TeamProjectReferenceCollection struct {
	Count *int                    `json:"count"`
	Value *[]TeamProjectReference `json:"value"`
}

type Time struct {
	Time time.Time
}

type WebApiTeam struct {
	Description *string    `json:"description,omitempty"`
	Id          *uuid.UUID `json:"id,omitempty"`
	Identity    *Identity  `json:"identity,omitempty"`
	IdentityUrl *string    `json:"identityUrl,omitempty"`
	Name        *string    `json:"name,omitempty"`
	ProjectId   *uuid.UUID `json:"projectId,omitempty"`
	ProjectName *string    `json:"projectName,omitempty"`
	Url         *string    `json:"url,omitempty"`
}

type WebApiTeamCollection struct {
	Count *int          `json:"count"`
	Value *[]WebApiTeam `json:"value"`
}

type WebApiTeamRef struct {
	Id   *uuid.UUID `json:"id,omitempty"`
	Name *string    `json:"name,omitempty"`
	Url  *string    `json:"url,omitempty"`
}
