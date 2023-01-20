package security

import (
	"context"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/security"
	"strings"
)

type IdentityPermissions struct {
	IdentityDescriptor string
	IdentityName       string
	Permissions        map[string]string
}

func CreateOrUpdateAccessControlEntry(ctx context.Context, namespaceId string, token string, permission *IdentityPermissions, securityClient *security.Client, graphClient *graph.Client) error {
	namespaces, namespacesErr := securityClient.GetSecurityNamespaces(ctx)
	if namespacesErr != nil {
		return namespacesErr
	}

	namespace := linq.From(*namespaces.Value).FirstWith(func(n interface{}) bool {
		return n.(security.SecurityNamespaceDescription).NamespaceId.String() == namespaceId
	}).(security.SecurityNamespaceDescription)

	accessControlEntry, err := getAccessControlEntry(ctx, &namespace, permission, securityClient, graphClient)
	if err != nil {
		return err
	}

	err = securityClient.SetAccessControlEntries(ctx, namespaceId, token, &[]security.AccessControlEntry{*accessControlEntry})
	if err != nil {
		return err
	}

	return nil
}

func CreateOrUpdateAccessControlList(ctx context.Context, namespaceId string, token string, permissions []*IdentityPermissions, securityClient *security.Client, graphClient *graph.Client) error {
	namespaces, namespacesErr := securityClient.GetSecurityNamespaces(ctx)
	if namespacesErr != nil {
		return namespacesErr
	}

	namespace := linq.From(*namespaces.Value).FirstWith(func(n interface{}) bool {
		return n.(security.SecurityNamespaceDescription).NamespaceId.String() == namespaceId
	}).(security.SecurityNamespaceDescription)

	accessControlList := &security.AccessControlList{
		AcesDictionary: &map[string]security.AccessControlEntry{},
		Token:          &token,
	}

	for _, permission := range permissions {
		accessControlEntry, err := getAccessControlEntry(ctx, &namespace, permission, securityClient, graphClient)
		if err != nil {
			return err
		}

		(*accessControlList.AcesDictionary)[permission.IdentityDescriptor] = *accessControlEntry
	}

	aclErr := securityClient.SetAccessControlLists(ctx, namespaceId, &[]security.AccessControlList{*accessControlList})
	if aclErr != nil {
		return aclErr
	}

	return nil
}

func ReadIdentityPermissions(ctx context.Context, namespaceId string, token string, securityClient *security.Client) ([]*IdentityPermissions, error) {
	accessControlLists, err := securityClient.GetAccessControlLists(ctx, namespaceId, token)
	if err != nil {
		return nil, err
	}

	if *accessControlLists.Count == 0 {
		return []*IdentityPermissions{}, errors.New("access control lists are empty")
	}

	namespaces, namespacesErr := securityClient.GetSecurityNamespaces(ctx)
	if namespacesErr != nil {
		return nil, namespacesErr
	}

	namespace := linq.From(*namespaces.Value).FirstWith(func(n interface{}) bool {
		return n.(security.SecurityNamespaceDescription).NamespaceId.String() == namespaceId
	}).(security.SecurityNamespaceDescription)

	var identityPermissions []*IdentityPermissions
	accessControlList := (*accessControlLists.Value)[0]

	for _, ace := range *accessControlList.AcesDictionary {
		identity, identityErr := securityClient.GetIdentityByDescriptor(ctx, *ace.Descriptor)
		if identityErr != nil {
			return nil, identityErr
		}

		permissions := map[string]string{}
		for _, action := range *namespace.Actions {
			if (*ace.Allow)&(*action.Bit) != 0 {
				permissions[*action.Name] = "allow"
			} else if (*ace.Deny)&(*action.Bit) != 0 {
				permissions[*action.Name] = "deny"
			} else {
				permissions[*action.Name] = "notset"
			}
		}

		identityName := identity.ProviderDisplayName
		properties := identity.Properties.(map[string]interface{})
		schemaClassName := properties["SchemaClassName"].(map[string]interface{})["$value"].(string)
		switch schemaClassName {
		case "Group":
			if strings.HasPrefix(*identity.SubjectDescriptor, "aadgp") {
				name := properties["Account"].(map[string]interface{})["$value"].(string)
				if mail, ok := properties["Mail"]; ok {
					name = mail.(map[string]interface{})["$value"].(string)
				}
				identityName = &name
			}
		case "User":
			account := properties["Account"].(map[string]interface{})["$value"].(string)
			identityName = &account
		default:
			return nil, errors.New(fmt.Sprintf("Unknown schema class name '%s'.", schemaClassName))
		}

		identityPermissions = append(identityPermissions, &IdentityPermissions{
			IdentityDescriptor: *ace.Descriptor,
			IdentityName:       *identityName,
			Permissions:        permissions,
		})
	}

	return identityPermissions, nil
}

// Private Methods

func getAccessControlEntry(ctx context.Context, namespace *security.SecurityNamespaceDescription, permission *IdentityPermissions, securityClient *security.Client, graphClient *graph.Client) (*security.AccessControlEntry, error) {
	if permission.IdentityDescriptor == "" {
		identity, identityErr := graphClient.GetIdentityPickerIdentity(ctx, permission.IdentityName)
		if identityErr != nil {
			return nil, identityErr
		}

		subjectDescriptor := identity.SubjectDescriptor
		if subjectDescriptor == nil {
			switch *identity.EntityType {
			case "Group":
				group, err := graphClient.CreateGroupByOriginId(ctx, *identity.OriginId)
				if err != nil {
					return nil, err
				}

				subjectDescriptor = group.Descriptor
			case "User":
				user, err := graphClient.CreateUserByOriginId(ctx, *identity.OriginId)
				if err != nil {
					return nil, err
				}

				subjectDescriptor = user.Descriptor
			default:
				return nil, errors.New(fmt.Sprintf("Unknown entity type '%s'", *identity.EntityType))
			}
		}

		identity2, err := securityClient.GetIdentityBySubjectDescriptor(ctx, *subjectDescriptor)
		if err != nil {
			return nil, err
		}

		permission.IdentityDescriptor = *identity2.Descriptor
	}

	allow := 0
	deny := 0

	for key, value := range permission.Permissions {
		action := linq.From(*namespace.Actions).FirstWith(func(a interface{}) bool {
			return strings.EqualFold(*a.(security.ActionDefinition).Name, key)
		}).(security.ActionDefinition)

		if strings.EqualFold("deny", value) {
			allow = allow &^ *action.Bit
			deny = deny | *action.Bit
		} else if strings.EqualFold("allow", value) {
			deny = deny &^ *action.Bit
			allow = allow | *action.Bit
		} else if strings.EqualFold("notset", value) {
			allow = allow &^ *action.Bit
			deny = deny &^ *action.Bit
		}
	}

	return &security.AccessControlEntry{
		Descriptor: &permission.IdentityDescriptor,
		Allow:      &allow,
		Deny:       &deny,
	}, nil
}
