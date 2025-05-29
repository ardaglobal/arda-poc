echo "== Initial balance of charlie"
arda-pocd q bank balances charlie

echo ""
echo "== Initial list of properties"
arda-pocd q property property-all

echo ""
echo "== Register property"
arda-pocd tx property register-property "4 sun st" dubai 1000000 --owners bob,alice,charlie --shares 10,20,70 --from ERES -y

echo ""
echo "== Transfer shares"
arda-pocd tx property transfer-shares "4 sun st" charlie 10 fred,george 5,5 --from ERES -y

echo ""
echo "== List of properties"
arda-pocd q property property-all

echo ""
echo "== Balance of charlie"
arda-pocd q bank balances charlie
