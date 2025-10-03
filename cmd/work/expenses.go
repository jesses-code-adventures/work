package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/jesses-code-adventures/work/internal/models"
	"github.com/jesses-code-adventures/work/internal/service"
)

func newExpensesCmd(timesheetService *service.TimesheetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expenses",
		Short: "Create, update and list expenses",
		Long:  "Commands for managing expenses, including listing expenses and their amounts.",
	}

	cmd.AddCommand(newExpensesCreateCmd(timesheetService))
	cmd.AddCommand(newExpensesListCmd(timesheetService))
	cmd.AddCommand(newExpensesUpdateCmd(timesheetService))

	return cmd
}

func newExpensesCreateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var amount float64
	var expenseDate, reference, client string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new expense",
		Long:  "Create an expense with a given amount, date, and optional reference and client",
		Args:  cobra.NoArgs,
	}

	cmd.Flags().Float64VarP(&amount, "amount", "a", 0.0, "Amount of the expense (required)")
	cmd.Flags().StringVarP(&expenseDate, "date", "d", "", "Date of the expense (YYYY-MM-DD, defaults to today)")
	cmd.Flags().StringVarP(&reference, "reference", "r", "", "Reference or description for the expense")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Client name to associate with the expense")

	cmd.MarkFlagRequired("amount")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if amount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}

		// Parse expense date
		var parsedDate time.Time
		var err error
		if expenseDate == "" {
			parsedDate = time.Now()
		} else {
			parsedDate, err = time.Parse("2006-01-02", expenseDate)
			if err != nil {
				return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
			}
		}

		// Get client ID if client name provided
		var clientID *string
		if client != "" {
			clientModel, err := timesheetService.GetClientByName(ctx, client)
			if err != nil {
				return fmt.Errorf("failed to find client '%s': %w", client, err)
			}
			clientID = &clientModel.ID
		}

		// Create reference pointer
		var refPtr *string
		if reference != "" {
			refPtr = &reference
		}

		expense, err := timesheetService.CreateExpense(ctx, decimal.NewFromFloat(amount), parsedDate, refPtr, clientID, nil)
		if err != nil {
			return fmt.Errorf("failed to create expense: %w", err)
		}

		fmt.Printf("Created expense: %s\n", expense.ID)
		timesheetService.DisplayExpense(ctx, expense)

		return nil
	}

	return cmd
}

func newExpensesListCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var verbose bool
	var client string
	var fromDate, toDate string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List expenses",
		Long:  "Display a list of expenses with optional filtering by client and date range.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var expenses []*models.Expense
			var err error

			// Determine which list method to use based on flags
			if client != "" && fromDate != "" && toDate != "" {
				// Client and date range
				startDate, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					return fmt.Errorf("invalid from date format, use YYYY-MM-DD: %w", err)
				}
				endDate, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					return fmt.Errorf("invalid to date format, use YYYY-MM-DD: %w", err)
				}
				expenses, err = timesheetService.ListExpensesByClientAndDateRange(ctx, client, startDate, endDate)
			} else if client != "" {
				// Client only
				expenses, err = timesheetService.ListExpensesByClient(ctx, client)
			} else if fromDate != "" && toDate != "" {
				// Date range only
				startDate, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					return fmt.Errorf("invalid from date format, use YYYY-MM-DD: %w", err)
				}
				endDate, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					return fmt.Errorf("invalid to date format, use YYYY-MM-DD: %w", err)
				}
				expenses, err = timesheetService.ListExpensesByDateRange(ctx, startDate, endDate)
			} else {
				// All expenses
				expenses, err = timesheetService.ListExpenses(ctx)
			}

			if err != nil {
				return fmt.Errorf("failed to list expenses: %w", err)
			}

			if len(expenses) == 0 {
				fmt.Println("No expenses found.")
				return nil
			}

			fmt.Printf("Found %d expense(s):\n\n", len(expenses))

			for _, expense := range expenses {
				if verbose {
					timesheetService.DisplayExpense(ctx, expense)
					fmt.Println()
				} else {
					fmt.Printf("%s - %s - %s",
						expense.ExpenseDate.Format("2006-01-02"),
						timesheetService.FormatBillableAmount(expense.Amount),
						expense.ID)

					if expense.Reference != nil && *expense.Reference != "" {
						fmt.Printf(" - %s", *expense.Reference)
					}

					if expense.ClientID != nil {
						client, err := timesheetService.GetClientByID(ctx, *expense.ClientID)
						if err == nil {
							fmt.Printf(" - %s", client.Name)
						}
					}

					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed expense information")
	cmd.Flags().StringVarP(&client, "client", "c", "", "Filter by client name")
	cmd.Flags().StringVar(&fromDate, "from", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "Filter to date (YYYY-MM-DD)")

	return cmd
}

func newExpensesUpdateCmd(timesheetService *service.TimesheetService) *cobra.Command {
	var amount float64
	var expenseDate, reference, client string

	cmd := &cobra.Command{
		Use:   "update <expense-id>",
		Short: "Update an expense",
		Long:  "Update attributes of an expense, such as amount, date, reference, or client.",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().Float64VarP(&amount, "amount", "a", 0.0, "New amount for the expense")
	cmd.Flags().StringVarP(&expenseDate, "date", "d", "", "New date for the expense (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&reference, "reference", "r", "", "New reference for the expense")
	cmd.Flags().StringVarP(&client, "client", "c", "", "New client name for the expense")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		expenseID := args[0]

		// Build update parameters
		var amountPtr *decimal.Decimal
		var datePtr *time.Time
		var refPtr *string
		var clientPtr *string

		if amount > 0 {
			amt := decimal.NewFromFloat(amount)
			amountPtr = &amt
		}

		if expenseDate != "" {
			parsedDate, err := time.Parse("2006-01-02", expenseDate)
			if err != nil {
				return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
			}
			datePtr = &parsedDate
		}

		if cmd.Flags().Changed("reference") {
			refPtr = &reference
		}

		if cmd.Flags().Changed("client") {
			clientPtr = &client
		}

		updatedExpense, err := timesheetService.UpdateExpense(ctx, expenseID, amountPtr, datePtr, refPtr, clientPtr, nil)
		if err != nil {
			return fmt.Errorf("failed to update expense: %w", err)
		}

		fmt.Printf("Updated expense '%s'\nNew state:\n", updatedExpense.ID)
		timesheetService.DisplayExpense(ctx, updatedExpense)

		return nil
	}

	return cmd
}
