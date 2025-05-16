package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/x/property/types"
)

func TestGetSetProperty(t *testing.T) {
	k, ctx := keepertest.PropertyKeeper(t)

	// Create a test property
	testProperty := types.Property{
		Index:   "property1",
		Address: "123 Main St",
		Region:  "Test Region",
		Value:   1000,
		Owners:  map[string]uint64{"cosmos1abcdefg": 100},
	}

	// Test that getting a non-existent property returns false
	_, found := k.GetProperty(ctx, testProperty.Index)
	require.False(t, found)

	// Test setting property
	k.SetProperty(ctx, testProperty)

	// Test getting the property we just set
	retrievedProperty, found := k.GetProperty(ctx, testProperty.Index)
	require.True(t, found)
	require.Equal(t, testProperty.Address, retrievedProperty.Address)
	require.Equal(t, testProperty.Region, retrievedProperty.Region)
	require.Equal(t, testProperty.Value, retrievedProperty.Value)
	require.Equal(t, testProperty.Owners, retrievedProperty.Owners)
}

func TestGetAllProperties(t *testing.T) {
	k, ctx := keepertest.PropertyKeeper(t)

	// Create multiple test properties
	testProperties := []types.Property{
		{
			Index:   "property1",
			Address: "123 Main St",
			Region:  "Test Region 1",
			Value:   1000,
			Owners:  map[string]uint64{"cosmos1abcdefg": 100},
		},
		{
			Index:   "property2",
			Address: "456 Oak Ave",
			Region:  "Test Region 2",
			Value:   2000,
			Owners:  map[string]uint64{"cosmos1hijklmn": 200},
		},
		{
			Index:   "property3",
			Address: "789 Pine Blvd",
			Region:  "Test Region 3",
			Value:   3000,
			Owners:  map[string]uint64{"cosmos1opqrstu": 300},
		},
	}

	// Set all test properties
	for _, prop := range testProperties {
		k.SetProperty(ctx, prop)
	}

	// Get all properties
	retrievedProperties, err := k.GetAllProperties(ctx)
	require.NoError(t, err)
	require.Len(t, retrievedProperties, len(testProperties))

	// Create a map for easier lookup
	propertyMap := make(map[string]types.Property)
	for _, prop := range retrievedProperties {
		propertyMap[prop.Index] = prop
	}

	// Verify all properties were retrieved correctly
	for _, expected := range testProperties {
		retrieved, found := propertyMap[expected.Index]
		require.True(t, found)
		require.Equal(t, expected.Address, retrieved.Address)
		require.Equal(t, expected.Region, retrieved.Region)
		require.Equal(t, expected.Value, retrieved.Value)
		require.Equal(t, expected.Owners, retrieved.Owners)
	}
}
