package graph

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
