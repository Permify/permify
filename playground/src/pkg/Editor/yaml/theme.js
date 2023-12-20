function Theme() {
    let bg = window.getComputedStyle(document.documentElement).getPropertyValue('--background-base').trim()
    return {
        base: 'vs-dark',
        inherit: true,
        colors: {
            "editor.background": bg,
        }
    }
}

export default Theme
