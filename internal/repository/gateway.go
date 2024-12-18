package repository

import (
	"database/sql"
	"fmt"
	"payment-gateway/internal/models"
	"time"
)

// GatewayRepository defines methods for retrieving gateways
type GatewayRepository interface {
	GetGatewaysByCountryAndCurrency(countryId, currency string) ([]*models.Gateway, error)
	GetGateways() ([]models.Gateway, error)
	CreateGateway(gateway models.Gateway) error
	GetSupportedCountriesByGateway(gatewayID int) ([]models.Country, error)
}

// GatewayRepositoryImpl is the concrete implementation of GatewayRepository
type GatewayRepositoryImpl struct {
	db *sql.DB
}

// NewGatewayRepository creates a new instance of GatewayRepository.
func NewGatewayRepository(db *sql.DB) *GatewayRepositoryImpl {
	return &GatewayRepositoryImpl{db: db}
}

func (g *GatewayRepositoryImpl) GetGateways() ([]models.Gateway, error) {
	rows, err := g.db.Query(`SELECT id, name, data_format_supported, created_at, updated_at FROM gateways`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways: %v", err)
	}
	defer rows.Close()

	var gateways []models.Gateway
	for rows.Next() {
		var gateway models.Gateway
		if err := rows.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported, &gateway.CreatedAt, &gateway.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gateways, nil
}

// GetGatewayByCurrency fetches all suitable gateways based on the currency.
//
//	Note: The first gateway of the list will be the latest created gateway
func (g *GatewayRepositoryImpl) GetGatewaysByCountryAndCurrency(countryId, currency string) ([]*models.Gateway, error) {
	query := `
		SELECT g.id, g.name, g.data_format_supported
		FROM gateways g
		JOIN gateway_countries gc ON g.id = gc.gateway_id
		JOIN countries c ON gc.country_id = c.id
		WHERE c.currency = $1
		and c.id = $2::numeric
		order by g.created_at desc;`

	rows, err := g.db.Query(query, currency, countryId)
	if err != nil {
		return nil, fmt.Errorf("couldn't find appropriate gateways: %v", err)
	}
	defer rows.Close()

	var gateways []*models.Gateway
	for rows.Next() {
		var gateway models.Gateway
		if err := rows.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, &gateway)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gateways, nil
}

func (g *GatewayRepositoryImpl) CreateGateway(gateway models.Gateway) error {
	query := `INSERT INTO gateways (name, data_format_supported, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := g.db.QueryRow(query, gateway.Name, gateway.DataFormatSupported, time.Now(), time.Now()).Scan(&gateway.ID)
	if err != nil {
		return fmt.Errorf("failed to insert gateway: %v", err)
	}
	return nil
}

func (g *GatewayRepositoryImpl) GetSupportedCountriesByGateway(gatewayID int) ([]models.Country, error) {
	query := `
		SELECT c.id AS country_id, c.name AS country_name
		FROM countries c
		JOIN gateway_countries gc ON c.id = gc.country_id
		WHERE gc.gateway_id = $1
		ORDER BY c.name
	`

	rows, err := g.db.Query(query, gatewayID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries for gateway %d: %v", gatewayID, err)
	}
	defer rows.Close()

	var countries []models.Country
	for rows.Next() {
		var country models.Country
		if err := rows.Scan(&country.ID, &country.Name); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		countries = append(countries, country)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %v", err)
	}

	return countries, nil
}
