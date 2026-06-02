# Loom Video Script

## 0:00 - Intro

Hi, this is Alvin De Castro. This is my practical assessment for the AI Web Developer role. I built a simple AI-powered freight quote assistant using a Go backend and a React frontend.

## 0:15 - Product idea

The goal is to help an operations or customer service user intake freight quote requests, calculate a demo estimate, identify missing details, and generate a customer-ready response.

A key design decision: the LLM does not calculate the freight price. The quote engine is deterministic and explainable. AI is used for summarization, missing-field review, and customer communication.

## 0:40 - Tech stack

The backend is Go using the standard net/http package. The frontend is React with Vite. The app has an optional OpenAI integration. If no API key is configured, it falls back to a local rule-based assistant so the demo still works.

## 1:00 - Walkthrough

Here is the quote form. I can enter origin, destination, shipment type, weight, dimensions, service level, accessorials, and notes.

I will load a demo quote: Vancouver to Calgary, two pallets, 700 pounds total, liftgate and residential delivery.

When I generate the quote, the app returns an estimated range, a cost breakdown, chargeable weight, confidence level, and a status.

## 1:40 - AI workflow

Here is the AI-assisted summary. It explains why the quote has certain charges and what the operations team should verify before sending it.

Here is the customer-ready response. The user can copy it and send it as a first draft.

## 2:10 - Messy request parser

Freight requests often arrive as messy customer emails. I added a parser where a user can paste a plain-English request. The app extracts route, pallets, dimensions, weight, accessorials, and service level into the form.

## 2:40 - Engineering tradeoffs

This is intentionally an MVP. It uses in-memory quote history and demo pricing logic. For production, I would add carrier API integrations, postal-code distance calculation, persistent storage, authentication, audit logs, admin rate tables, and CRM integration.

## 3:05 - Closing

This project shows how I would approach an AI logistics tool: keep critical business rules deterministic, use AI for acceleration and communication, and build a workflow that operations teams can actually use.
