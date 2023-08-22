# Disclaimer

This fork is a tentative work in progress to fork **Terraform** into another project called **OpenTF**.

**Hashicorp** is not affiliated in any way with this project and does not endorse it. Any mention of **Terraform** or **Hashicorp** or any other trademark belonging to **Harshicorp** are accidental legacy artifacts that will be gradually phased out of the project.

The original **Terraform** project maintained by **Hashicorp** can be found here: https://github.com/hashicorp/terraform

We are grateful to **Hashicorp** for having created and stewarded the open-source version of **Terraform** (which this project is based) on as well as its community for so long.

# OpenTF

- Website: https://opentf.org/
- Forums: https://opentf.org/
- Documentation: https://opentf.org/
- Tutorials: https://opentf.org/

OpenTF is a tool for building, changing, and versioning infrastructure safely and efficiently. OpenTF can manage existing and popular service providers as well as custom in-house solutions.

The key features of OpenTF are:

- **Infrastructure as Code**: Infrastructure is described using a high-level configuration syntax. This allows a blueprint of your datacenter to be versioned and treated as you would any other code. Additionally, infrastructure can be shared and re-used.

- **Execution Plans**: OpenTF has a "planning" step where it generates an execution plan. The execution plan shows what OpenTF will do when you call apply. This lets you avoid any surprises when OpenTF manipulates infrastructure.

- **Resource Graph**: OpenTF builds a graph of all your resources, and parallelizes the creation and modification of any non-dependent resources. Because of this, OpenTF builds infrastructure as efficiently as possible, and operators get insight into dependencies in their infrastructure.

- **Change Automation**: Complex changesets can be applied to your infrastructure with minimal human interaction. With the previously mentioned execution plan and resource graph, you know exactly what OpenTF will change and in what order, avoiding many possible human errors.

For more information, refer to the [What is OpenTF?](https://opentf.org/) page on the OpenTF website.

## Getting Started & Documentation

Documentation is available on the [OpenTF website](https://opentf.org/):

- [Introduction](https://opentf.org/)
- [Documentation](https://opentf.org/)

If you're new to OpenTF and want to get started creating infrastructure, please check out our [Getting Started guides](https://opentf.org/) on OpenTF's learning platform. There are also [additional guides](https://opentf.org/) to continue your learning.

## Developing OpenTF

This repository contains only OpenTF core, which includes the command line interface and the main graph engine. Providers are implemented as plugins, and OpenTF can automatically download providers that are published on [the OpenTF Registry](https://opentf.org/). The providers are developed by various individuals and organizations. For more information, see [Extending OpenTF](https://opentf.org/).

- To learn more about compiling OpenTF and contributing suggested changes, refer to [the contributing guide](.github/CONTRIBUTING.md).

- To learn more about how we handle bug reports, refer to the [bug triage guide](./BUGPROCESS.md).

- To learn how to contribute to the OpenTF documentation in this repository, refer to the [OpenTF Documentation README](/website/README.md).

## License

[Mozilla Public License v2.0](https://github.com/Magnitus-/opentf/blob/main/LICENSE)
