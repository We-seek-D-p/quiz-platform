import Aura from '@primevue/themes/aura'
import { definePreset } from '@primevue/themes'

export const AuraIndigoPreset = definePreset(Aura, {
  semantic: {
    formField: {
      paddingX: '1rem',
      paddingY: '0.75rem',
      borderRadius: '0.5rem',
    },
    primary: {
      50: '{indigo.50}',
      100: '{indigo.100}',
      200: '{indigo.200}',
      300: '{indigo.300}',
      400: '{indigo.400}',
      500: '{indigo.500}',
      600: '{indigo.600}',
      700: '{indigo.700}',
      800: '{indigo.800}',
      900: '{indigo.900}',
      950: '{indigo.950}',
    },
    colorScheme: {
      light: {
        surface: {
          0: '#ffffff',
          50: '#f1f4f6',
          100: '#e2e8ee',
          200: '#c6d1dd',
          300: '#a9bbcb',
          400: '#8da4ba',
          500: '#708da9',
          600: '#5a7187',
          700: '#435565',
          800: '#2d3844',
          900: '#161c22',
          950: '#0d1117',
        },
      },
      dark: {
        surface: {
          0: '#ffffff',
          50: '#e8e9e9',
          100: '#d2d2d4',
          200: '#bbbcbe',
          300: '#a5a5a9',
          400: '#8e8f93',
          500: '#77787d',
          600: '#616268',
          700: '#4a4b52',
          800: '#34343d',
          900: '#1d1e27',
          950: '#14151d',
        },
      },
    },
  },
  components: {
    button: {
      root: {
        paddingX: '1.125rem',
        paddingY: '0.625rem',
        borderRadius: '{form.field.border.radius}',
        iconOnlyWidth: '2.5rem',
        sm: {
          fontSize: '0.875rem',
          paddingX: '0.75rem',
          paddingY: '0.625rem',
          iconOnlyWidth: '2.25rem',
        },
        label: {
          fontWeight: '500',
        },
        primary: {
          background: '{primary.600}',
          hoverBackground: '{primary.700}',
          activeBackground: '{primary.800}',
          borderColor: '{primary.600}',
          hoverBorderColor: '{primary.700}',
          activeBorderColor: '{primary.800}',
        },
      },
    },
    inputtext: {
      root: {
        paddingX: '{form.field.padding.x}',
        paddingY: '{form.field.padding.y}',
        borderRadius: '{form.field.border.radius}',
      },
    },
    textarea: {
      root: {
        paddingX: '{form.field.padding.x}',
        paddingY: '{form.field.padding.y}',
        borderRadius: '{form.field.border.radius}',
      },
    },
    select: {
      root: {
        paddingX: '{form.field.padding.x}',
        paddingY: '{form.field.padding.y}',
        borderRadius: '{form.field.border.radius}',
      },
    },
    card: {
      root: {
        borderRadius: '0.5rem',
      },
      body: {
        padding: '1.25rem',
        gap: '1rem',
      },
    },
    dialog: {
      root: {
        borderRadius: '0.5rem',
      },
      header: {
        padding: '1.25rem 1.25rem 1rem 1.25rem',
      },
      content: {
        padding: '0 1.25rem 1.25rem 1.25rem',
      },
      footer: {
        padding: '0 1.25rem 1.25rem 1.25rem',
        gap: '0.5rem',
      },
    },
  },
})
