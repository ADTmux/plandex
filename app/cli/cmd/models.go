package cmd

import (
	"fmt"
	"os"
	"plandex/api"
	"plandex/auth"
	"plandex/lib"
	"plandex/term"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/plandex/plandex/shared"
	"github.com/spf13/cobra"
)

var customModelsOnly bool

func init() {
	RootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(listAvailableModelsCmd)
	modelsCmd.AddCommand(createCustomModelCmd)

	listAvailableModelsCmd.Flags().BoolVarP(&customModelsOnly, "custom", "c", false, "List custom models only")
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Show model settings",
	Run:   models,
}

var listAvailableModelsCmd = &cobra.Command{
	Use:   "available",
	Short: "List all available models",
	Run:   listAvailableModels,
}

var createCustomModelCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom model",
	Run:   createCustomModel,
}

func createCustomModel(cmd *cobra.Command, args []string) {
	model := &shared.CustomModel{}

	modelName, err := term.GetUserStringInput("Enter model name:")
	if err != nil {
		term.OutputErrorAndExit("Error reading model name: %v", err)
		return
	}
	model.ModelName = modelName

	provider, err := term.GetUserStringInput("Enter provider:")
	if err != nil {
		term.OutputErrorAndExit("Error reading provider: %v", err)
		return
	}
	model.Provider = provider

	baseUrl, err := term.GetUserStringInput("Enter base URL:")
	if err != nil {
		term.OutputErrorAndExit("Error reading base URL: %v", err)
		return
	}
	model.BaseUrl = baseUrl

	maxTokensStr, err := term.GetUserStringInput("Enter max tokens:")
	if err != nil {
		term.OutputErrorAndExit("Error reading max tokens: %v", err)
		return
	}
	maxTokens, err := strconv.Atoi(maxTokensStr)
	if err != nil {
		term.OutputErrorAndExit("Invalid number for max tokens: %v", err)
		return
	}
	model.MaxTokens = maxTokens

	apiKeyEnvVar, err := term.GetUserStringInput("Enter API key environment variable:")
	if err != nil {
		term.OutputErrorAndExit("Error reading API key environment variable: %v", err)
		return
	}
	model.ApiKeyEnvVar = apiKeyEnvVar

	description, err := term.GetUserStringInput("Enter description (optional):")
	if err != nil {
		term.OutputErrorAndExit("Error reading description: %v", err)
		return
	}
	model.Description = description

	term.StartSpinner("Creating model...")
	err = api.Client.CreateCustomModel(model)
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error creating model: %v", err)
		return
	}

	fmt.Println("Model created successfully.")
}

func models(cmd *cobra.Command, args []string) {

	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	term.StartSpinner("")
	settings, err := api.Client.GetSettings(lib.CurrentPlanId, lib.CurrentBranch)
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error getting settings: %v", err)
		return
	}

	modelSet := settings.ModelSet
	if modelSet == nil {
		modelSet = shared.DefaultModelSet
	}

	color.New(color.Bold, term.ColorHiCyan).Println("🎛️  Current Model Set")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(true)
	table.SetColWidth(64)
	table.SetHeader([]string{modelSet.Name})
	table.Append([]string{modelSet.Description})
	table.Render()
	fmt.Println()

	color.New(color.Bold, term.ColorHiCyan).Println("🤖 Models")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Role", "Provider", "Model", "Temperature", "Top P"})

	addModelRow := func(role string, config shared.ModelRoleConfig) {
		table.Append([]string{
			role,
			string(config.BaseModelConfig.Provider),
			config.BaseModelConfig.ModelName,
			fmt.Sprintf("%.1f", config.Temperature),
			fmt.Sprintf("%.1f", config.TopP),
		})
	}

	addModelRow(string(shared.ModelRolePlanner), modelSet.Planner.ModelRoleConfig)
	addModelRow(string(shared.ModelRolePlanSummary), modelSet.PlanSummary)
	addModelRow(string(shared.ModelRoleBuilder), modelSet.Builder)
	addModelRow(string(shared.ModelRoleName), modelSet.Namer)
	addModelRow(string(shared.ModelRoleCommitMsg), modelSet.CommitMsg)
	addModelRow(string(shared.ModelRoleExecStatus), modelSet.ExecStatus)
	table.Render()

	fmt.Println()

	color.New(color.Bold, term.ColorHiCyan).Println("🧠 Planner Defaults")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Max Tokens", "Max Convo Tokens", "Reserved Output Tokens"})
	table.Append([]string{
		fmt.Sprintf("%d", modelSet.Planner.BaseModelConfig.MaxTokens),
		fmt.Sprintf("%d", modelSet.Planner.MaxConvoTokens),
		fmt.Sprintf("%d", modelSet.Planner.ReservedOutputTokens),
	})
	table.Render()
	fmt.Println()

	color.New(color.Bold, term.ColorHiCyan).Println("⚙️  Planner Overrides")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Name", "Value"})
	if settings.ModelOverrides.MaxTokens == nil {
		table.Append([]string{"Max Tokens", "no override"})
	} else {
		table.Append([]string{"Max Tokens", fmt.Sprintf("%d", *settings.ModelOverrides.MaxTokens)})
	}
	if settings.ModelOverrides.MaxConvoTokens == nil {
		table.Append([]string{"Max Convo Tokens", "no override"})
	} else {
		table.Append([]string{"Max Convo Tokens", fmt.Sprintf("%d", *settings.ModelOverrides.MaxConvoTokens)})
	}
	if settings.ModelOverrides.ReservedOutputTokens == nil {
		table.Append([]string{"Reserved Output Tokens", "no override"})
	} else {
		table.Append([]string{"Reserved Output Tokens", fmt.Sprintf("%d", *settings.ModelOverrides.ReservedOutputTokens)})
	}
	table.Render()
	fmt.Println()

	term.PrintCmds("", "set-model", "models available")
}

func listAvailableModels(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	if lib.CurrentPlanId == "" {
		fmt.Println("🤷‍♂️ No current plan")
		return
	}

	term.StartSpinner("")

	customModels, err := api.Client.ListCustomModels()

	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error fetching custom models: %v", err)
		return
	}

	if !customModelsOnly {
		color.New(color.Bold, term.ColorHiCyan).Println("🏠 Built-in Models")
		builtIn := shared.AvailableModels
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"Provider", "Name", "🪙", "🔑", "Url"})
		for _, model := range builtIn {
			table.Append([]string{string(model.Provider), model.ModelName, strconv.Itoa(model.MaxTokens), model.ApiKeyEnvVar, model.BaseUrl})
		}
		table.Render()
		fmt.Println()
	}

	if len(customModels) > 0 {
		color.New(color.Bold, term.ColorHiCyan).Println("🛠️ Custom Models")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"#", "Provider", "Name", "🪙", "🔑", "Url"})
		for i, model := range customModels {
			table.Append([]string{fmt.Sprintf("%d", i+1), model.ModelName, string(model.Provider), strconv.Itoa(model.MaxTokens), model.ApiKeyEnvVar, model.BaseUrl})
		}
		table.Render()
	} else if customModelsOnly {
		fmt.Println("🤷‍♂️ No custom models")
		fmt.Println()
	}

	if customModelsOnly {
		term.PrintCmds("", "models", "set-model")
	} else {
		term.PrintCmds("", "models available --custom", "models", "set-model")
	}
}
func createCustomModel(cmd *cobra.Command, args []string) {
	model := &shared.CustomModel{}

	term.StartSpinner("Creating model...")
	err := api.Client.CreateCustomModel(model)
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error creating model: %v", err)
		return
	}

	fmt.Println("Model created successfully.")
var deleteCustomModelCmd = &cobra.Command{
	Use:   "delete [name-or-index]",
	Short: "Delete a custom model by name or index",
	Args:  cobra.MaximumNArgs(1),
	Run:   deleteCustomModel,
}

func deleteCustomModel(cmd *cobra.Command, args []string) {
	var models []shared.CustomModel
	var err error

	term.StartSpinner("Fetching custom models...")
	models, err = api.Client.ListCustomModels()
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error fetching custom models: %v", err)
		return
	}

	if len(models) == 0 {
		fmt.Println("No custom models available to delete.")
		return
	}

	var modelToDelete *shared.CustomModel

	if len(args) == 1 {
		input := args[0]
		// Try to parse input as index
		index, err := strconv.Atoi(input)
		if err == nil && index > 0 && index <= len(models) {
			modelToDelete = &models[index-1]
		} else {
			// Search by name
			for _, m := range models {
				if m.ModelName == input {
					modelToDelete = &m
					break
				}
			}
		}
	}

	if modelToDelete == nil {
		fmt.Println("Select a model to delete:")
		for i, model := range models {
			fmt.Printf("%d: %s\n", i+1, model.ModelName)
		}
		var selectedIndex int
		fmt.Scanln(&selectedIndex)
		if selectedIndex < 1 || selectedIndex > len(models) {
			fmt.Println("Invalid selection.")
			return
		}
		modelToDelete = &models[selectedIndex-1]
	}

	term.StartSpinner(fmt.Sprintf("Deleting model '%s'...", modelToDelete.ModelName))
	err = api.Client.DeleteCustomModel(modelToDelete.Id)
	term.StopSpinner()

	if err != nil {
		term.OutputErrorAndExit("Error deleting custom model: %v", err)
		return
	}

	fmt.Printf("Custom model '%s' deleted successfully.\n", modelToDelete.ModelName)
}



