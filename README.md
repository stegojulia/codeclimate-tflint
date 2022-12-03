# Code Climate tflint Engine

[![Code Climate](https://codeclimate.com/github/ST-Apps/codeclimate-tflint/badges/gpa.svg)](https://codeclimate.com/github/ST-Apps/codeclimate-tflint)

`codeclimate-tflint` is a Code Climate engine that wraps [TFLint](https://github.com/terraform-linters/tflint). You can run it on your command line using the Code Climate CLI, or on our hosted analysis platform.

TFLint is a pluggable [Terraform](https://www.terraform.io/) Linter.

> This engine is based on TFLint v0.43.0

### Installation

1. If you haven't already, [install the Code Climate CLI](https://github.com/codeclimate/codeclimate).
2. Run `codeclimate engines:enable tflint`. This command both installs the engine and enables it in your `.codeclimate.yml` file.
3. You're ready to analyze! Browse into your project's folder and run `codeclimate analyze`.

### Configuration

By default, TFLint will look for a `.tflint.hcl` file in the root of
your project. Optionally configure Code Climate to look at a different path:

```yml
plugins:
  tflint:
    enabled: true
    config:
      config: optional/path/to/.tflint.hcl
```

In the same way you can set all the options supported by TFLint (more details [here](https://github.com/terraform-linters/tflint#usage)):

```yml
# .codeclimate.yml
plugins:
  tflint:
    enabled: true
    config:
      config: "" # --config=FILE
      ignore_module: [ # --ignore-module=SOURCE
        "..."
      ]
      enable_rule: [ # --enable-rule=RULE_NAME
        "..."
      ]
      disable_rule: [ # --disable-rule=RULE_NAME
        "..."
      ]
      only: [ # --only=RULE_NAME
        "..."
      ]
      enable_plugin: [ # --enable-plugin=PLUGIN_NAME
        "..."
      ]
      var_file: "" # --var-file=FILE
      var: [ # --var='foo=bar'
        "..."
      ]
      module: true # --module 
```

### Need help?

For help with TFLint, [check out their documentation](https://github.com/terraform-linters/tflint).

If you're running into a Code Climate issue, first look over this project's [GitHub Issues](https://github.com/ST-Apps/codeclimate-tflint/issues), as your question may have already been covered. If not, [go ahead and open a support ticket with us](https://codeclimate.com/help).