package models

// TransactionType enumerates the supported transaction types.
type TransactionType string

const (
	TransactionIncome      TransactionType = "INCOME"
	TransactionExpense     TransactionType = "EXPENSE"
	TransactionSaving      TransactionType = "SAVING"
	TransactionInstallment TransactionType = "INSTALLMENTS"
)

// Currency enumerates the supported currencies.
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyUYU Currency = "UYU"
	CurrencyEUR Currency = "EUR"
)

// Month enumerates the supported calendar months.
type Month string

const (
	MonthJanuary   Month = "JANUARY"
	MonthFebruary  Month = "FEBRUARY"
	MonthMarch     Month = "MARCH"
	MonthApril     Month = "APRIL"
	MonthMay       Month = "MAY"
	MonthJune      Month = "JUNE"
	MonthJuly      Month = "JULY"
	MonthAugust    Month = "AUGUST"
	MonthSeptember Month = "SEPTEMBER"
	MonthOctober   Month = "OCTOBER"
	MonthNovember  Month = "NOVEMBER"
	MonthDecember  Month = "DECEMBER"
)
