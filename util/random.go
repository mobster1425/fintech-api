package util

import (

	"fmt"
	"math/rand"
	"strings"
	"time"
)



var (
	// For generating random strings
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Possible values for fields
	Roles        = []string{"customer", "merchant"}
//	Roles        = []string{string(db.UserRoleCustomer), string(db.UserRoleMerchant)}
	statuses     = []string{"active", "banned", "inactive"}

transaction_types = []string{"TRANSFER",
"REQUEST",
"REDEEM",
"PAYMENT",
"PAYMENT_VOUCHER",
"WITHDRAW",
"DEPOSIT"}

transaction_status = []string{ "INIT",
"PROCESSING",
"PENDING",
"SUCCESS",
"CANCELED",
"REJECTED"}


voucher_types= []string{ "FIXED",
"PERCENT"}

voucher_status=[]string{ "AVAILABLE",
"UNAVAILABLE"}


)



// const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}



// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}


// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}



// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}


// RandomBool generates a random boolean value
func RandomBool() bool {
	return rand.Intn(2) == 0
}

// RandomRole generates a random role (customer or merchant)
func RandomRole() string {
	return Roles[rand.Intn(len(Roles))]
}

// RandomStatus generates a random status (active, banned, or inactive)
func RandomStatus() string {
	return statuses[rand.Intn(len(statuses))]
}




func RandomTransactionType() string {
	return transaction_types[rand.Intn(len(transaction_types))]
}


func RandomTransactionStatus() string {
	return transaction_status[rand.Intn(len(transaction_status))]
}


func RandomVoucherStatus() string {
	return voucher_status[rand.Intn(len(voucher_status))]
}


func RandomVoucherType() string {
	return voucher_types[rand.Intn(len(voucher_types))]
}

// Helper function to generate a random alphanumeric code
func GenerateRandomCode() string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    code := make([]byte, 10)
    for i := range code {
        code[i] = charset[rand.Intn(len(charset))]
    }
    return string(code)
}


var chargeFeeValue = 0.02;

func CalculateAmount(amount float64, chargeForSender bool) (float64 , float64, float64) {
	sendAmount := 0.0
	receiveAmount := 0.0

	chargeFee := amount * chargeFeeValue
	if !chargeForSender {
		// charge for receiver
		sendAmount = amount
		receiveAmount = amount - chargeFee
	} else {
		// charge for sender
		sendAmount = amount + chargeFee
		receiveAmount = amount
	}

	

	return sendAmount, receiveAmount,chargeFee
}
