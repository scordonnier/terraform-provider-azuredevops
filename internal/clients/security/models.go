package security

import "github.com/google/uuid"

type AccessControlEntry struct {
	Allow        *int                    `json:"allow,omitempty"`
	Deny         *int                    `json:"deny,omitempty"`
	Descriptor   *string                 `json:"descriptor,omitempty"`
	ExtendedInfo *AceExtendedInformation `json:"extendedInfo,omitempty"`
}

type AccessControlEntryCollection struct {
	Count *int                  `json:"count"`
	Value *[]AccessControlEntry `json:"value"`
}

type AccessControlList struct {
	AcesDictionary      *map[string]AccessControlEntry `json:"acesDictionary,omitempty"`
	IncludeExtendedInfo *bool                          `json:"includeExtendedInfo,omitempty"`
	InheritPermissions  *bool                          `json:"inheritPermissions,omitempty"`
	Token               *string                        `json:"token,omitempty"`
}

type AccessControlListCollection struct {
	Count *int                 `json:"count"`
	Value *[]AccessControlList `json:"value"`
}

type AceExtendedInformation struct {
	EffectiveAllow *int `json:"effectiveAllow,omitempty"`
	EffectiveDeny  *int `json:"effectiveDeny,omitempty"`
	InheritedAllow *int `json:"inheritedAllow,omitempty"`
	InheritedDeny  *int `json:"inheritedDeny,omitempty"`
}

type ActionDefinition struct {
	Bit         *int       `json:"bit,omitempty"`
	DisplayName *string    `json:"displayName,omitempty"`
	Name        *string    `json:"name,omitempty"`
	NamespaceId *uuid.UUID `json:"namespaceId,omitempty"`
}

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

type IdentityCollection struct {
	Count *int        `json:"count"`
	Value *[]Identity `json:"value"`
}

type SecurityNamespaceDescription struct {
	Actions            *[]ActionDefinition `json:"actions,omitempty"`
	DataspaceCategory  *string             `json:"dataspaceCategory,omitempty"`
	DisplayName        *string             `json:"displayName,omitempty"`
	ElementLength      *int                `json:"elementLength,omitempty"`
	ExtensionType      *string             `json:"extensionType,omitempty"`
	IsRemotable        *bool               `json:"isRemotable,omitempty"`
	Name               *string             `json:"name,omitempty"`
	NamespaceId        *uuid.UUID          `json:"namespaceId,omitempty"`
	ReadPermission     *int                `json:"readPermission,omitempty"`
	SeparatorValue     *string             `json:"separatorValue,omitempty"`
	StructureValue     *int                `json:"structureValue,omitempty"`
	SystemBitMask      *int                `json:"systemBitMask,omitempty"`
	UseTokenTranslator *bool               `json:"useTokenTranslator,omitempty"`
	WritePermission    *int                `json:"writePermission,omitempty"`
}

type SecurityNamespacesCollection struct {
	Count *int                            `json:"count"`
	Value *[]SecurityNamespaceDescription `json:"value"`
}

type SetAccessControlEntriesArgs struct {
	AccessControlEntries *[]AccessControlEntry `json:"accessControlEntries,omitempty"`
	Merge                *bool                 `json:"merge,omitempty"`
	Token                *string               `json:"token,omitempty"`
}
