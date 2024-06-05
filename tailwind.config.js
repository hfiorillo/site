/** @type {import('tailwindcss').Config} */
module.exports = {
 	content: [ "./**/*.html", "./**/*.templ", "./**/*.go", ],
	safelist: [],
	plugins: [
		require("@tailwindcss/typography"), require("daisyui")
	],
	daisyui: {
		themes: ["lofi", "dark", "cupcake", "synthwave"],
	}
}