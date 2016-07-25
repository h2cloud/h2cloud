package outapi

// Used for providing a united connector.

import (
    . "definition/configinfo"
)

var DefaultConnector=ConnectbyAuth(
    KEYSTONE_USERNAME,
    KEYSTONE_PASSWORD,
    KEYSTONE_TENANT)
