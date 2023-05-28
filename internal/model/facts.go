package model

// Facts is stored into System table as JSONB
type Facts struct {
	BiosVendor            string `json:"bios_vendor"`
	BiosVersion           string `json:"bios_version"`
	BiosReleaseDate       string `json:"bios_release_date"`
	BiosRevision          string `json:"bios_revision"`
	FirmwareRevision      string `json:"firmware_revision"`
	SystemManufacturer    string `json:"system_manufacturer"`
	SystemProductName     string `json:"system_product_name"`
	SystemVersion         string `json:"system_version"`
	SystemSerialNumber    string `json:"system_serial_number"`
	SystemUUID            string `json:"system_uuid"`
	SystemSkuNumber       string `json:"system_sku_number"`
	SystemFamily          string `json:"system_family"`
	BaseboardManufacturer string `json:"baseboard_manufacturer"`
	BaseboardProductName  string `json:"baseboard_product_name"`
	BaseboardVersion      string `json:"baseboard_version"`
	BaseboardSerialNumber string `json:"baseboard_serial_number"`
	BaseboardAssetTag     string `json:"baseboard_asset_tag"`
	ChassisManufacturer   string `json:"chassis_manufacturer"`
	ChassisType           string `json:"chassis_type"`
	ChassisVersion        string `json:"chassis_version"`
	ChassisSerialNumber   string `json:"chassis_serial_number"`
	ChassisAssetTag       string `json:"chassis_asset_tag"`
	ProcessorFamily       string `json:"processor_family"`
	ProcessorManufacturer string `json:"processor_manufacturer"`
	ProcessorVersion      string `json:"processor_version"`
	ProcessorFrequency    string `json:"processor_frequency"`
}
