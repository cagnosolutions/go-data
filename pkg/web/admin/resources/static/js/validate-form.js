// validation documentation: https://github.com/horprogs/Just-validate

// login form validation
const loginValidation = new window.JustValidate('#login-form', {
    errorFieldCssClass: 'is-invalid',
    successFieldCssClass: 'is-valid',
    lockForm: false,
}).onSuccess((e) => {
    loginValidation.form.submit();
});

loginValidation
    .addField('#username', [
        {rule: 'required', errorMessage: 'Username is required!'},
        {rule: 'email', errorMessage: 'Username must be a valid email address!'}
    ])
    .addField("#password", [
        {rule: 'required', errorMessage: 'Password is required!'},
        {rule: 'password'},
    ]);