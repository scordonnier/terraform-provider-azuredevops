package graph

import "github.com/google/uuid"

type GraphDescriptorResult struct {
	Links interface{} `json:"_links,omitempty"`
	Value *string     `json:"value,omitempty"`
}

type GraphGroup struct {
	Description      *string     `json:"description,omitempty"`
	Descriptor       *string     `json:"descriptor,omitempty"`
	DisplayName      *string     `json:"displayName,omitempty"`
	Domain           *string     `json:"domain,omitempty"`
	LegacyDescriptor *string     `json:"legacyDescriptor,omitempty"`
	Links            interface{} `json:"_links,omitempty"`
	MailAddress      *string     `json:"mailAddress,omitempty"`
	PrincipalName    *string     `json:"principalName,omitempty"`
	Origin           *string     `json:"origin,omitempty"`
	OriginId         *string     `json:"originId,omitempty"`
	SubjectKind      *string     `json:"subjectKind,omitempty"`
	Url              *string     `json:"url,omitempty"`
}

type GraphGroupCollection struct {
	Count *int          `json:"count"`
	Value *[]GraphGroup `json:"value"`
}

type GraphGroupVstsCreationContext struct {
	CrossProject         *bool      `json:"crossProject,omitempty"`
	Description          *string    `json:"description,omitempty"`
	Descriptor           *string    `json:"descriptor,omitempty"`
	DisplayName          *string    `json:"displayName,omitempty"`
	RestrictedVisibility *bool      `json:"restrictedVisibility,omitempty"`
	SpecialGroupType     *string    `json:"specialGroupType,omitempty"`
	StorageKey           *uuid.UUID `json:"storageKey,omitempty"`
}

type GraphUser struct {
	Descriptor        *string     `json:"descriptor,omitempty"`
	DisplayName       *string     `json:"displayName,omitempty"`
	DirectoryAlias    *string     `json:"directoryAlias,omitempty"`
	Domain            *string     `json:"domain,omitempty"`
	IsDeletedInOrigin *bool       `json:"isDeletedInOrigin,omitempty"`
	LegacyDescriptor  *string     `json:"legacyDescriptor,omitempty"`
	Links             interface{} `json:"_links,omitempty"`
	MailAddress       *string     `json:"mailAddress,omitempty"`
	MetaType          *string     `json:"metaType,omitempty"`
	Origin            *string     `json:"origin,omitempty"`
	OriginId          *string     `json:"originId,omitempty"`
	PrincipalName     *string     `json:"principalName,omitempty"`
	SubjectKind       *string     `json:"subjectKind,omitempty"`
	Url               *string     `json:"url,omitempty"`
}

type GraphUserCollection struct {
	Count *int         `json:"count"`
	Value *[]GraphUser `json:"value"`
}

type IdentityPickerIdentity struct {
	Active                     *bool   `json:"active,omitempty"`
	Department                 *string `json:"department,omitempty"`
	Description                *string `json:"description,omitempty"`
	DisplayName                *string `json:"displayName,omitempty"`
	EntityId                   *string `json:"entityId,omitempty"`
	EntityType                 *string `json:"entityType,omitempty"`
	Guest                      *bool   `json:"guest,omitempty"`
	IsMru                      *bool   `json:"isMru,omitempty"`
	JobTitle                   *string `json:"jobTitle,omitempty"`
	LocalDirectory             *string `json:"localDirectory,omitempty"`
	LocalId                    *string `json:"localId,omitempty"`
	Mail                       *string `json:"mail,omitempty"`
	MailNickname               *string `json:"mailNickname,omitempty"`
	OriginDirectory            *string `json:"originDirectory,omitempty"`
	OriginId                   *string `json:"originId,omitempty"`
	PhysicalDeliveryOfficeName *string `json:"physicalDeliveryOfficeName,omitempty"`
	SamAccountName             *string `json:"samAccountName,omitempty"`
	ScopeName                  *string `json:"scopeName,omitempty"`
	SignInAddress              *string `json:"signInAddress,omitempty"`
	SubjectDescriptor          *string `json:"subjectDescriptor,omitempty"`
	Surname                    *string `json:"surname,omitempty"`
	TelephoneNumber            *string `json:"telephoneNumber,omitempty"`
}

type IdentityPickerOptions struct {
	MaxResults int `json:"MaxResults,omitempty"`
	MinResults int `json:"MinResults,omitempty"`
}

type IdentityPickerRequest struct {
	IdentityTypes   *[]string              `json:"identityTypes,omitempty"`
	OperationScopes *[]string              `json:"operationScopes,omitempty"`
	Options         *IdentityPickerOptions `json:"options,omitempty"`
	Properties      *[]string              `json:"properties,omitempty"`
	Query           *string                `json:"query,omitempty"`
}

type IdentityPickerResponse struct {
	Results *[]IdentityPickerResult `json:"results,omitempty"`
}

type IdentityPickerResult struct {
	Identities  *[]IdentityPickerIdentity `json:"identities,omitempty"`
	PagingToken *string                   `json:"pagingToken,omitempty"`
	QueryToken  *string                   `json:"queryToken,omitempty"`
}
