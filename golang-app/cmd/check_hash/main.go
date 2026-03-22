package main

import (
"fmt"
"golang.org/x/crypto/bcrypt"
)

func main() {
hash := "$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK"
passwords := []string{"12345678", "password", "password123", "admin123"}
for _, p := range passwords {
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
if err == nil {
fmt.Printf("MATCH: '%s'\n", p)
} else {
fmt.Printf("NO: '%s'\n", p)
}
}
}
