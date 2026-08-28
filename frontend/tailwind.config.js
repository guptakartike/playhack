/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          dark: '#0B0B12',
          surface: '#14141E',
          card: '#1A1A26',
          border: '#2A2A3C',
          orange: '#F5793A',
          'orange-hover': '#E06728',
          text: '#F1F5F9',
          muted: '#94A3B8',
        }
      }
    },
  },
  plugins: [],
}
