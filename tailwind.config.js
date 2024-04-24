/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/**/*.{tmpl,gohtml,html,js}"],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}

