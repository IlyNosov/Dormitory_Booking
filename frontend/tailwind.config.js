export default {
    content: ["./index.html", "./src/**/*.{ts,tsx}"],
    theme: {
        extend: {
            colors: {
                bg:               '#F4F1EB',
                surface:          '#FDFCFA',
                panel:            '#EDE9E0',
                border:           '#D6CEBC',
                primary:          '#5B7A3F',
                'primary-dark':   '#3A5428',
                'primary-light':  '#A8C484',
                secondary:        '#9B6B3E',
                accent:           '#C49A5E',
                success:          '#4E8B5F',
                error:            '#A64B3A',
                warn:             '#B8893A',
                text:             '#2C3520',
                'text-muted':     '#7A8569',
                'text-secondary': '#5C6B4A',
            },
            fontFamily: {
                display: ['Cormorant Garant', 'Georgia', 'serif'],
            },
        },
    },
    plugins: [],
};
