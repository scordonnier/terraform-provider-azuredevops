package security

import (
	"context"
	"errors"
	"github.com/ahmetb/go-linq/v3"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/security"
	"strings"
)

type IdentityPermissions struct {
	IdentityDescriptor string
	IdentityName       string
	IdentityType       string
	Permissions        map[string]string
}

func CreateOrUpdateIdentityPermissions(ctx context.Context, namespaceId string, token string, permissions []*IdentityPermissions, client *security.Client) error {
	namespaces, namespacesErr := client.GetSecurityNamespaces(ctx)
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
		if permission.IdentityDescriptor == "" {
			identity, identityErr := client.GetIdentity(ctx, permission.IdentityName, permission.IdentityType)
			if identityErr != nil {
				return identityErr
			}

			permission.IdentityDescriptor = *identity.Descriptor
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

		(*accessControlList.AcesDictionary)[permission.IdentityDescriptor] = security.AccessControlEntry{
			Descriptor: &permission.IdentityDescriptor,
			Allow:      &allow,
			Deny:       &deny,
		}
	}

	aclErr := client.SetAccessControlLists(ctx, namespaceId, &[]security.AccessControlList{*accessControlList})
	if aclErr != nil {
		return aclErr
	}

	return nil
}

func ReadIdentityPermissions(ctx context.Context, namespaceId string, token string, client *security.Client) ([]*IdentityPermissions, error) {
	accessControlLists, err := client.GetAccessControlLists(ctx, namespaceId, token)
	if err != nil {
		return nil, err
	}

	if *accessControlLists.Count == 0 {
		return []*IdentityPermissions{}, errors.New("access control lists are empty")
	}

	namespaces, namespacesErr := client.GetSecurityNamespaces(ctx)
	if namespacesErr != nil {
		return nil, namespacesErr
	}

	namespace := linq.From(*namespaces.Value).FirstWith(func(n interface{}) bool {
		return n.(security.SecurityNamespaceDescription).NamespaceId.String() == namespaceId
	}).(security.SecurityNamespaceDescription)

	var identityPermissions []*IdentityPermissions
	accessControlList := (*accessControlLists.Value)[0]

	for _, value := range *accessControlList.AcesDictionary {
		identity, identityErr := client.GetIdentityByDescriptor(ctx, *value.Descriptor)
		if identityErr != nil {
			return nil, identityErr
		}

		permissions := map[string]string{}
		for _, action := range *namespace.Actions {
			if (*value.Allow)&(*action.Bit) != 0 {
				permissions[*action.Name] = "allow"
			} else if (*value.Deny)&(*action.Bit) != 0 {
				permissions[*action.Name] = "deny"
			} else {
				permissions[*action.Name] = "notset"
			}
		}

		identityName := *identity.ProviderDisplayName
		identityType := "group"
		if strings.HasPrefix(*value.Descriptor, "Microsoft.IdentityModel.Claims.ClaimsIdentity") {
			identityName = identity.Properties.(map[string]interface{})["Account"].(map[string]interface{})["$value"].(string)
			identityType = "user"
		}

		identityPermissions = append(identityPermissions, &IdentityPermissions{
			IdentityDescriptor: *value.Descriptor,
			IdentityName:       identityName,
			IdentityType:       identityType,
			Permissions:        permissions,
		})
	}

	return identityPermissions, nil
}
