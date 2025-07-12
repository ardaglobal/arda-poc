package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
)

// ValidatePropertyExists checks if a property exists and returns it
func (k Keeper) ValidatePropertyExists(ctx sdk.Context, propertyId string) (propertytypes.Property, error) {
	property, found := k.propertyKeeper.GetProperty(ctx, propertyId)
	if !found {
		return propertytypes.Property{}, fmt.Errorf("property with ID %s not found", propertyId)
	}
	return property, nil
}

// ValidatePropertyOwnership checks if the given address is an owner of the property
func (k Keeper) ValidatePropertyOwnership(ctx sdk.Context, propertyId string, ownerAddr string) error {
	property, err := k.ValidatePropertyExists(ctx, propertyId)
	if err != nil {
		return err
	}

	// Convert owners to map for easier lookup
	ownerMap := k.propertyKeeper.ConvertPropertyOwnersToMap(property)
	
	if _, isOwner := ownerMap[ownerAddr]; !isOwner {
		return fmt.Errorf("address %s is not an owner of property %s", ownerAddr, propertyId)
	}
	
	return nil
}

// GetPropertyOwners returns all owners of a property with their shares
func (k Keeper) GetPropertyOwners(ctx sdk.Context, propertyId string) (map[string]uint64, error) {
	property, err := k.ValidatePropertyExists(ctx, propertyId)
	if err != nil {
		return nil, err
	}

	return k.propertyKeeper.ConvertPropertyOwnersToMap(property), nil
}

// GetPropertyValue returns the value of a property
func (k Keeper) GetPropertyValue(ctx sdk.Context, propertyId string) (uint64, error) {
	property, err := k.ValidatePropertyExists(ctx, propertyId)
	if err != nil {
		return 0, err
	}

	return property.Value, nil
}

// IsPropertyAvailableForRent checks if a property is available for rent
// This is a business logic function that can be customized based on requirements
func (k Keeper) IsPropertyAvailableForRent(ctx sdk.Context, propertyId string) (bool, error) {
	property, err := k.ValidatePropertyExists(ctx, propertyId)
	if err != nil {
		return false, err
	}

	// Check if property already has an active tenant
	// This is a simplified check - in a real implementation, you might want to:
	// 1. Check if there are any active leases for this property
	// 2. Check property-specific rental availability flags
	// 3. Check property zoning and legal restrictions
	
	if property.TenantId != "" {
		return false, fmt.Errorf("property %s already has an active tenant: %s", propertyId, property.TenantId)
	}

	return true, nil
}