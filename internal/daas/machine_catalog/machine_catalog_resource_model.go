// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MachineCatalogResourceModel maps the resource schema data.
type MachineCatalogResourceModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	IsPowerManaged         types.Bool   `tfsdk:"is_power_managed"`
	IsRemotePc             types.Bool   `tfsdk:"is_remote_pc"`
	AllocationType         types.String `tfsdk:"allocation_type"`
	SessionSupport         types.String `tfsdk:"session_support"`
	Zone                   types.String `tfsdk:"zone"`
	VdaUpgradeType         types.String `tfsdk:"vda_upgrade_type"`
	ProvisioningType       types.String `tfsdk:"provisioning_type"`
	ProvisioningScheme     types.Object `tfsdk:"provisioning_scheme"` // ProvisioningSchemeModel
	MachineAccounts        types.List   `tfsdk:"machine_accounts"`    // List[MachineAccountsModel]
	RemotePcOus            types.List   `tfsdk:"remote_pc_ous"`       // List[RemotePcOuModel]
	MinimumFunctionalLevel types.String `tfsdk:"minimum_functional_level"`
	Scopes                 types.Set    `tfsdk:"scopes"` //Set[String]
}

type MachineAccountsModel struct {
	Hypervisor types.String `tfsdk:"hypervisor"`
	Machines   types.List   `tfsdk:"machines"` // List[MachineCatalogMachineModel]
}

func (MachineAccountsModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"hypervisor": schema.StringAttribute{
				Description: "The Id of the hypervisor in which the machines reside. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"machines": schema.ListNestedAttribute{
				Description:  "Machines to add to the catalog",
				Required:     true,
				NestedObject: MachineCatalogMachineModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (MachineAccountsModel) GetAttributes() map[string]schema.Attribute {
	return MachineAccountsModel{}.GetSchema().Attributes
}

type MachineCatalogMachineModel struct {
	MachineAccount    types.String `tfsdk:"machine_account"`
	MachineName       types.String `tfsdk:"machine_name"`
	Region            types.String `tfsdk:"region"`
	ResourceGroupName types.String `tfsdk:"resource_group_name"`
	ProjectName       types.String `tfsdk:"project_name"`
	AvailabilityZone  types.String `tfsdk:"availability_zone"`
	Datacenter        types.String `tfsdk:"datacenter"`
	Cluster           types.String `tfsdk:"cluster"`
	Host              types.String `tfsdk:"host"`
}

func (MachineCatalogMachineModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"machine_account": schema.StringAttribute{
				Description: "The Computer AD Account for the machine. Must be in the format DOMAIN\\MACHINE.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.SamRegex), "must be in the format DOMAIN\\MACHINE"),
				},
			},
			"machine_name": schema.StringAttribute{
				Description: "The name of the machine. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "**[Azure, GCP: Required]** The region in which the machine resides. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"resource_group_name": schema.StringAttribute{
				Description: "**[Azure: Required]** The resource group in which the machine resides. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"project_name": schema.StringAttribute{
				Description: "**[GCP: Required]** The project name in which the machine resides. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"availability_zone": schema.StringAttribute{
				Description: "**[AWS: Required]** The availability zone in which the machine resides. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"datacenter": schema.StringAttribute{
				Description: "**[vSphere: Required]** The datacenter in which the machine resides. Required only if `is_power_managed = true`",
				Optional:    true,
			},
			"cluster": schema.StringAttribute{
				Description: "**[vSphere: Optional]** The cluster in which the machine resides. To be used only if `is_power_managed = true`",
				Optional:    true,
			},
			"host": schema.StringAttribute{
				Description: "**[vSphere, SCVMM: Required]** For vSphere, this is the IP address or FQDN of the host in which the machine resides. For SCVMM, this is the name of the host in which the machine resides. Required only if `is_power_managed = true`",
				Optional:    true,
			},
		},
	}
}

func (MachineCatalogMachineModel) GetAttributes() map[string]schema.Attribute {
	return MachineCatalogMachineModel{}.GetSchema().Attributes
}

// ProvisioningSchemeModel maps the nested provisioning scheme resource schema data.
type ProvisioningSchemeModel struct {
	Hypervisor                  types.String `tfsdk:"hypervisor"`
	HypervisorResourcePool      types.String `tfsdk:"hypervisor_resource_pool"`
	AzureMachineConfig          types.Object `tfsdk:"azure_machine_config"`     // AzureMachineConfigModel
	AwsMachineConfig            types.Object `tfsdk:"aws_machine_config"`       // AwsMachineConfigModel
	GcpMachineConfig            types.Object `tfsdk:"gcp_machine_config"`       // GcpMachineConfigModel
	VsphereMachineConfig        types.Object `tfsdk:"vsphere_machine_config"`   // VsphereMachineConfigModel
	XenserverMachineConfig      types.Object `tfsdk:"xenserver_machine_config"` // XenserverMachineConfigModel
	NutanixMachineConfig        types.Object `tfsdk:"nutanix_machine_config"`   // NutanixMachineConfigModel
	NumTotalMachines            types.Int64  `tfsdk:"number_of_total_machines"`
	NetworkMapping              types.List   `tfsdk:"network_mapping"`    // List[NetworkMappingModel]
	AvailabilityZones           types.List   `tfsdk:"availability_zones"` // List[string]
	IdentityType                types.String `tfsdk:"identity_type"`
	MachineDomainIdentity       types.Object `tfsdk:"machine_domain_identity"`        // MachineDomainIdentityModel
	MachineAccountCreationRules types.Object `tfsdk:"machine_account_creation_rules"` // MachineAccountCreationRulesModel
	CustomProperties            types.List   `tfsdk:"custom_properties"`              // List[CustomPropertyModel]
}

func (ProvisioningSchemeModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Machine catalog provisioning scheme. Required when `provisioning_type = MCS` or `provisioning_type = PVS_STREAMING`.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"hypervisor": schema.StringAttribute{
				Description: "Id of the hypervisor for creating the machines. Required only if using power managed machines.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"hypervisor_resource_pool": schema.StringAttribute{
				Description: "Id of the hypervisor resource pool that will be used for provisioning operations.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"azure_machine_config":     AzureMachineConfigModel{}.GetSchema(),
			"aws_machine_config":       AwsMachineConfigModel{}.GetSchema(),
			"gcp_machine_config":       GcpMachineConfigModel{}.GetSchema(),
			"vsphere_machine_config":   VsphereMachineConfigModel{}.GetSchema(),
			"xenserver_machine_config": XenserverMachineConfigModel{}.GetSchema(),
			"nutanix_machine_config":   NutanixMachineConfigModel{}.GetSchema(),
			"machine_domain_identity":  MachineDomainIdentityModel{}.GetSchema(),
			"number_of_total_machines": schema.Int64Attribute{
				Description: "Number of VDA machines allocated in the catalog.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"network_mapping": schema.ListNestedAttribute{
				Description: "Specifies how the attached NICs are mapped to networks. If this parameter is omitted, provisioned VMs are created with a single NIC, which is mapped to the default network in the hypervisor resource pool.  If this parameter is supplied, machines are created with the number of NICs specified in the map, and each NIC is attached to the specified network." + "<br />" +
					"Required when `provisioning_scheme.identity_type` is `AzureAD`.",
				Optional:     true,
				NestedObject: NetworkMappingModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"availability_zones": schema.ListAttribute{
				Description: "The Availability Zones for provisioning virtual machines.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"identity_type": schema.StringAttribute{
				Description: "The identity type of the machines to be created. Supported values are`ActiveDirectory`, `AzureAD`, and `HybridAzureAD`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY),
						string(citrixorchestration.IDENTITYTYPE_AZURE_AD),
						string(citrixorchestration.IDENTITYTYPE_HYBRID_AZURE_AD),
						string(citrixorchestration.IDENTITYTYPE_WORKGROUP),
					),
					validators.AlsoRequiresOnValues(
						[]string{
							string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY),
						},
						path.MatchRelative().AtParent().AtName("machine_domain_identity"),
					),
					validators.AlsoRequiresOnValues(
						[]string{
							string(citrixorchestration.IDENTITYTYPE_HYBRID_AZURE_AD),
						},
						path.MatchRelative().AtParent().AtName("machine_domain_identity"),
					),
					validators.AlsoRequiresOnValues(
						[]string{
							string(citrixorchestration.IDENTITYTYPE_AZURE_AD),
						},
						path.MatchRelative().AtParent().AtName("azure_machine_config"),
						path.MatchRelative().AtParent().AtName("azure_machine_config").AtName("machine_profile"),
						path.MatchRelative().AtParent().AtName("network_mapping"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_account_creation_rules": MachineAccountCreationRulesModel{}.GetSchema(),
			"custom_properties": schema.ListNestedAttribute{
				Description:  "**This is an advanced feature. Use with caution.** Custom properties to be set for the machine catalog. For properties that are already supported as a terraform configuration field, please use terraform field instead.",
				Optional:     true,
				NestedObject: CustomPropertyModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (ProvisioningSchemeModel) GetAttributes() map[string]schema.Attribute {
	return ProvisioningSchemeModel{}.GetSchema().Attributes
}

type CustomPropertyModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (CustomPropertyModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the custom property.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Value of the custom property.",
				Required:    true,
			},
		},
	}
}

func (CustomPropertyModel) GetAttributes() map[string]schema.Attribute {
	return CustomPropertyModel{}.GetSchema().Attributes
}

type MachineDomainIdentityModel struct {
	Domain                 types.String `tfsdk:"domain"`
	Ou                     types.String `tfsdk:"domain_ou"`
	ServiceAccount         types.String `tfsdk:"service_account"`
	ServiceAccountPassword types.String `tfsdk:"service_account_password"`
}

func (MachineDomainIdentityModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The domain identity for machines in the machine catalog." + "<br />" +
			"Required when identity_type is set to `ActiveDirectory`",
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The AD domain name for the pool. Specify this in FQDN format; for example, MyDomain.com.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.DomainFqdnRegex), "must be in FQDN format"),
				},
			},
			"domain_ou": schema.StringAttribute{
				Description: "The organization unit that computer accounts will be created into.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service_account": schema.StringAttribute{
				Description: "Service account for the domain. Only the username is required; do not include the domain name.",
				Required:    true,
			},
			"service_account_password": schema.StringAttribute{
				Description: "Service account password for the domain.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (MachineDomainIdentityModel) GetAttributes() map[string]schema.Attribute {
	return MachineDomainIdentityModel{}.GetSchema().Attributes
}

// MachineAccountCreationRulesModel maps the nested machine account creation rules resource schema data.
type MachineAccountCreationRulesModel struct {
	NamingScheme     types.String `tfsdk:"naming_scheme"`
	NamingSchemeType types.String `tfsdk:"naming_scheme_type"`
}

func (MachineAccountCreationRulesModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Rules specifying how Active Directory machine accounts should be created when machines are provisioned.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"naming_scheme": schema.StringAttribute{
				Description: "Defines the template name for AD accounts created in the identity pool.",
				Required:    true,
			},
			"naming_scheme_type": schema.StringAttribute{
				Description: "Type of naming scheme. This defines the format of the variable part of the AD account names that will be created. Choose between `Numeric`, `Alphabetic` and `Unicode`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedAccountNamingSchemeTypeEnumValues),
				},
			},
		},
	}
}

func (MachineAccountCreationRulesModel) GetAttributes() map[string]schema.Attribute {
	return MachineAccountCreationRulesModel{}.GetSchema().Attributes
}

// ensure NetworkMappingModel implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixorchestration.NetworkMapResponseModel] = NetworkMappingModel{}

// NetworkMappingModel maps the nested network mapping resource schema data.
type NetworkMappingModel struct {
	NetworkDevice types.String `tfsdk:"network_device"`
	Network       types.String `tfsdk:"network"`
}

func (n NetworkMappingModel) GetKey() string {
	return n.NetworkDevice.ValueString()
}

func (NetworkMappingModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"network_device": schema.StringAttribute{
				Description: "Name or Id of the network device.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("network"),
					}...),
				},
			},
			"network": schema.StringAttribute{
				Description: "The name of the virtual network that the device should be attached to. This must be a subnet within a Virtual Private Cloud item in the resource pool to which the Machine Catalog is associated." + "<br />" +
					"For AWS, please specify the network mask of the network you want to use within the VPC.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("network_device"),
					}...),
				},
			},
		},
	}
}

func (NetworkMappingModel) GetAttributes() map[string]schema.Attribute {
	return NetworkMappingModel{}.GetSchema().Attributes
}

// ensure RemotePcOuModel implements RefreshableListItemWithAttributes
var _ util.RefreshableListItemWithAttributes[citrixorchestration.RemotePCEnrollmentScopeResponseModel] = RemotePcOuModel{}

type RemotePcOuModel struct {
	IncludeSubFolders types.Bool   `tfsdk:"include_subfolders"`
	OUName            types.String `tfsdk:"ou_name"`
}

func (r RemotePcOuModel) GetKey() string {
	return r.OUName.ValueString()
}

func (RemotePcOuModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"include_subfolders": schema.BoolAttribute{
				Description: "Specify if subfolders should be included.",
				Required:    true,
			},
			"ou_name": schema.StringAttribute{
				Description: "Name of the OU.",
				Required:    true,
			},
		},
	}
}

func (RemotePcOuModel) GetAttributes() map[string]schema.Attribute {
	return RemotePcOuModel{}.GetSchema().Attributes
}

func (MachineCatalogResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a machine catalog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the machine catalog.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine catalog.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the machine catalog.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"is_power_managed": schema.BoolAttribute{
				Description: "Specify if the machines in the machine catalog will be power managed.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"is_remote_pc": schema.BoolAttribute{
				Description: "Specify if this catalog is for Remote PC access.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"allocation_type": schema.StringAttribute{
				Description: "Denotes how the machines in the catalog are allocated to a user. Choose between `Static` and `Random`. Allocation type should be `Random` when `session_support = MultiSession`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedAllocationTypeEnumValues),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"session_support": schema.StringAttribute{
				Description: "Session support type. Choose between `SingleSession` and `MultiSession`. Session support should be SingleSession when `is_remote_pc = true`.",
				Required:    true,
				Validators: []validator.String{
					util.GetValidatorFromEnum(citrixorchestration.AllowedSessionSupportEnumValues),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "Id of the zone the machine catalog is associated with.",
				Required:    true,
			},
			"vda_upgrade_type": schema.StringAttribute{
				Description: "Type of Vda Upgrade. Choose between LTSR and CR. When omitted, Vda Upgrade is disabled.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"LTSR",
						"CR",
					),
				},
			},
			"provisioning_type": schema.StringAttribute{
				Description: "Specifies how the machines are provisioned in the catalog.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.PROVISIONINGTYPE_MCS),
						string(citrixorchestration.PROVISIONINGTYPE_MANUAL),
						string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_accounts": schema.ListNestedAttribute{
				Description:  "Machine accounts to add to the catalog. Only to be used when using `provisioning_type = MANUAL`",
				Optional:     true,
				NestedObject: MachineAccountsModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"remote_pc_ous": schema.ListNestedAttribute{
				Description:  "Organizational Units to be included in the Remote PC machine catalog. Only to be used when `is_remote_pc = true`. For adding machines, use `machine_accounts`.",
				Optional:     true,
				NestedObject: RemotePcOuModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"minimum_functional_level": schema.StringAttribute{
				Description: "Specifies the minimum functional level for the VDA machines in the catalog. Defaults to `L7_20`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("L7_20"),
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(util.GetAllowedFunctionalLevelValues()...),
				},
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the machine catalog to be a part of.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"provisioning_scheme": ProvisioningSchemeModel{}.GetSchema(),
		},
	}
}

func (MachineCatalogResourceModel) GetAttributes() map[string]schema.Attribute {
	return MachineCatalogResourceModel{}.GetSchema().Attributes
}

func (r MachineCatalogResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixclient.CitrixDaasClient, catalog *citrixorchestration.MachineCatalogDetailResponseModel, connectionType *citrixorchestration.HypervisorConnectionType, machines *citrixorchestration.MachineResponseModelCollection, pluginId string) MachineCatalogResourceModel {
	// Machine Catalog Properties
	r.Id = types.StringValue(catalog.GetId())
	r.Name = types.StringValue(catalog.GetName())
	r.Description = types.StringValue(catalog.GetDescription())
	allocationType := catalog.GetAllocationType()
	r.AllocationType = types.StringValue(allocationTypeEnumToString(allocationType))
	sessionSupport := catalog.GetSessionSupport()
	r.SessionSupport = types.StringValue(string(sessionSupport))

	minimumFunctionalLevel := catalog.GetMinimumFunctionalLevel()
	r.MinimumFunctionalLevel = types.StringValue(string(minimumFunctionalLevel))

	catalogZone := catalog.GetZone()
	r.Zone = types.StringValue(catalogZone.GetId())

	if catalog.UpgradeInfo != nil {
		if *catalog.UpgradeInfo.UpgradeType != citrixorchestration.VDAUPGRADETYPE_NOT_SET || !r.VdaUpgradeType.IsNull() {
			r.VdaUpgradeType = types.StringValue(string(*catalog.UpgradeInfo.UpgradeType))
		}
	} else {
		r.VdaUpgradeType = types.StringNull()
	}

	provtype := catalog.GetProvisioningType()
	provScheme := catalog.GetProvisioningScheme()
	provSchemeType := provScheme.GetProvisioningSchemeType()

	if provSchemeType == "PVS" {
		// For PVS Streaming, provisioning type returned (MCS) is different from the value sent in schema (PVSStreaming)
		r.ProvisioningType = types.StringValue(string(citrixorchestration.PROVISIONINGTYPE_PVS_STREAMING))
	} else {
		r.ProvisioningType = types.StringValue(string(provtype))
	}
	if provtype == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		r.IsPowerManaged = types.BoolValue(catalog.GetIsPowerManaged())
	} else {
		r.IsPowerManaged = types.BoolNull()
	}

	if catalog.ProvisioningType == citrixorchestration.PROVISIONINGTYPE_MANUAL {
		// Handle machines
		r = r.updateCatalogWithMachines(ctx, diagnostics, client, machines)
	}

	r = r.updateCatalogWithRemotePcConfig(ctx, diagnostics, catalog)

	if catalog.ProvisioningScheme == nil {
		if attributesMap, err := util.AttributeMapFromObject(ProvisioningSchemeModel{}); err == nil {
			r.ProvisioningScheme = types.ObjectNull(attributesMap)
		} else {
			diagnostics.AddWarning("Error when creating null ProvisioningSchemeModel", err.Error())
		}
		return r
	}

	scopeIds := util.GetIdsForScopeObjects(catalog.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	// Provisioning Scheme Properties
	r = r.updateCatalogWithProvScheme(ctx, diagnostics, client, catalog, connectionType, pluginId, provScheme)

	return r
}

func (networkMapping NetworkMappingModel) RefreshListItem(_ context.Context, _ *diag.Diagnostics, nic citrixorchestration.NetworkMapResponseModel) util.ModelWithAttributes {
	networkMapping.NetworkDevice = types.StringValue(nic.GetDeviceId())
	network := nic.GetNetwork()
	segments := strings.Split(network.GetXDPath(), "\\")
	lastIndex := len(segments)

	networkName := (strings.Split(segments[lastIndex-1], "."))[0]
	matchAws := regexp.MustCompile(util.AwsNetworkNameRegex)
	if matchAws.MatchString(networkName) {
		/* For AWS Network, the XDPath looks like:
		* XDHyp:\\HostingUnits\\{resource pool}\\{availability zone}.availabilityzone\\{network ip}`/{prefix length} (vpc-{vpc-id}).network
		* The Network property should be set to {network ip}/{prefix length}
		 */
		networkName = strings.ReplaceAll(strings.Split((strings.Split(segments[lastIndex-1], ".network"))[0], " ")[0], "`/", "/")
	}
	networkMapping.Network = types.StringValue(networkName)
	return networkMapping
}
