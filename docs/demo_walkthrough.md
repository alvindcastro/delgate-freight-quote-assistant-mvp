# Demo Walkthrough

## Scenario

A customer asks for an LTL freight quote:

> Need to ship 2 pallets from Vancouver to Calgary. Each pallet is about 48x40x60 and 350 lbs. Customer needs liftgate and residential delivery before Friday.

## Steps

1. Open the frontend.
2. Paste the request into the messy request parser.
3. Click **Parse request**.
4. Confirm the extracted fields:
   - Origin: Vancouver, BC
   - Destination: Calgary, AB
   - Pallets: 2
   - Dimensions: 48 x 40 x 60 inches
   - Weight: 700 lbs total
   - Accessorials: liftgate, residential delivery
5. Click **Generate quote**.
6. Review:
   - Estimated quote range
   - Chargeable weight
   - Fuel surcharge
   - Accessorial fees
   - Confidence level
   - Manual review status
   - Missing information checklist
   - AI-generated operations summary
   - Customer-ready response
7. Use the **Copy customer response** button.

## Key point to mention

The quote calculation is deterministic and transparent. AI is used to improve the workflow around the quote, not to invent pricing.
