package utils

func BoolValue(b *bool, defaultVal bool) bool {
    if b != nil {
        return *b
    }
    return defaultVal
}

func UintValue(u *uint, defaultVal uint) uint {
    if u != nil {
        return *u
    }
    return defaultVal
}

func FloatValue(u *float64, defaultVal float64) float64 {
    if u != nil {
        return *u
    }
    return defaultVal
}


