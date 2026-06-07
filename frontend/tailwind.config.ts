import type { Config } from 'tailwindcss'

export default <Partial<Config>>{
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#1DBC60',
          50: '#e8f7ee',
          100: '#cceedd',
          200: '#99e0bb',
          300: '#66d399',
          400: '#33c677',
          500: '#1DBC60',
          600: '#17974d',
          700: '#12723a',
          800: '#0c4d28',
          900: '#062816',
        }
      }
    }
  }
}
