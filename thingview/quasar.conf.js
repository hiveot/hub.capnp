// is this used with vite, or only with the quasar cli?
module.exports = function (ctx) {
    return {
        supportTS: true,
        framework: {
            plugins: [
                'Notify', 'Dialog'
            ]
        }
    }
}