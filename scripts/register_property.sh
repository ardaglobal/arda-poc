#!/bin/bash

# Check if BOB is set
if [ -z "$BOB" ]; then
	echo "Error: BOB environment variable is not set"
	exit 1
fi

# Default values
ADDRESS="123 Test St"
REGION="Test Region"
VALUE="1000000"
OWNERS="$BOB=100"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
	case $1 in
	--address)
		ADDRESS="$2"
		shift 2
		;;
	--region)
		REGION="$2"
		shift 2
		;;
	--value)
		VALUE="$2"
		shift 2
		;;
	--owners)
		OWNERS="$2"
		shift 2
		;;
	*)
		echo "Unknown option: $1"
		exit 1
		;;
	esac
done

echo "Registering property:"
echo "Address: $ADDRESS"
echo "Region: $REGION"
echo "Value: $VALUE"
echo "Owners: $OWNERS"

# Display the command
echo -e "\nRunning command:"
echo "ardad tx property register-property \"$ADDRESS\" \"$REGION\" \"$VALUE\" --owners=\"$OWNERS\" --from=$BOB --home=.arda_data -y"

# Execute the command
ardad tx property register-property "$ADDRESS" "$REGION" "$VALUE" --owners="$OWNERS" --from=$BOB --home=.arda_data -y

# Wait a moment for the transaction to be processed
sleep 2

# Query the property
echo -e "\nQuerying property:"
ID=$(echo "$ADDRESS" | tr '[:upper:]' '[:lower:]' | tr -d ' ')
echo "ardad q property property --index=\"$ID\" --output=json"
ardad q property property --index="$ID" --output=json
