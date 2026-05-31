import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import pluginVue from "eslint-plugin-vue";
import vueParser from "vue-eslint-parser";

export default tseslint.config(
  { ignores: ["dist"] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ["**/*.{ts,vue}"],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
  },
  // Vue flat config: enables the SFC parser so <template> blocks lint correctly.
  ...pluginVue.configs["flat/recommended"],
  {
    // Parse *.vue with vue-eslint-parser, delegating <script lang="ts"> to the
    // TS parser so script blocks stay type-aware.
    files: ["**/*.vue"],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tseslint.parser,
        ecmaVersion: 2022,
        sourceType: "module",
      },
    },
    rules: {
      // Opinionated naming rule: would force renaming single-word SFCs like
      // Logo. Kept off rather than churning component names/imports repo-wide.
      "vue/multi-word-component-names": "off",
      // Purely stylistic template-formatting rules from flat/recommended. They
      // are the formatter's job (Prettier) and enabling them would reformat
      // nearly every SFC. Disabled to keep lint signal meaningful instead of
      // drowning real issues under hundreds of whitespace warnings.
      "vue/max-attributes-per-line": "off",
      "vue/singleline-html-element-content-newline": "off",
      "vue/attributes-order": "off",
      "vue/html-self-closing": "off",
      "vue/html-indent": "off",
      "vue/html-closing-bracket-spacing": "off",
    },
  }
);
