/** @type {import('tailwindcss').Config} */
module.exports = {
 	content: [ "./**/*.html", "./**/*.templ", "./**/*.go", ],
	safelist: [],
	plugins: [
		require("@tailwindcss/typography"), require("daisyui")
	],
	daisyui: {
		themes: ["dark"]
	}
}