import eslintPluginReact from 'eslint-plugin-react';
import eslintPluginReactHooks from 'eslint-plugin-react-hooks';
import parser from '@typescript-eslint/parser';
import eslintPluginTS from '@typescript-eslint/eslint-plugin';

export default [
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    languageOptions: {
      parser,
      parserOptions: {
        ecmaVersion: 2020,
        sourceType: 'module',
        ecmaFeatures: { jsx: true }
      }
    },
    plugins: {
      '@typescript-eslint': eslintPluginTS,
      react: eslintPluginReact,
      'react-hooks': eslintPluginReactHooks
    },
    rules: {
      'react/prop-types': 'off',
      'react/react-in-jsx-scope': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off'
    }
  }
];
