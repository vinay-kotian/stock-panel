# Stock Panel - Go Options Information API

This project provides a simple HTTP API to collect and retrieve options information.

## Project Structure

- `cmd/stock-panel/main.go` — Main application entry point
- `internal/models/` — Data models (e.g., Stock struct)
- `internal/db/` — Database logic and initialization
- `internal/handlers/` — HTTP handler functions
- `internal/dashboard/` — Placeholder for future dashboard module
- `internal/login/` — Placeholder for future login/authentication module

## Features
- Collect options data via POST `/stocks`
- Retrieve all collected options data via GET `/stocks`

## Getting Started

### Prerequisites
- Go 1.18 or newer

### Running the Server

```bash
cd cmd/stock-panel
go run main.go
```

The server will start on `http://localhost:8080`.

### API Endpoints

#### POST /stocks
Collect options information.

**Request Body:**
```json
{
  "symbol": "AAPL240621C00150000",
  "underlying_symbol": "AAPL",
  "option_type": "CALL",
  "strike_price": 150.0,
  "expiry": "2024-06-21",
  "price": 5.25,
  "side": "BUY"
}
```

#### GET /stocks
Retrieve all collected options information.

**Response:**
```json
[
  {
    "symbol": "AAPL240621C00150000",
    "underlying_symbol": "AAPL",
    "option_type": "CALL",
    "strike_price": 150.0,
    "expiry": "2024-06-21",
    "price": 5.25,
    "side": "BUY",
    "timestamp": "2024-06-07T12:34:56.789Z"
  }
]
```

## Adding New Modules

To add new features (e.g., dashboard, login), create a new folder under `internal/` and add your code there. See the `internal/dashboard/` and `internal/login/` folders for placeholders.
