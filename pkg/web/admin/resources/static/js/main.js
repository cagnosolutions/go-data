function test() {
    alert("this is a test, this is only a test.");
}

// show and hide alerta
function hideAlertFade(id) {
    $(id).fadeOut(500);
}

function showAlertFade(id) {
    $(id).fadeIn(500);
}


/*
// Fetch all the forms we want to apply custom Bootstrap validation styles to
var forms = document.querySelectorAll('.needs-validation')
// Loop over them and prevent submission
Array.prototype.slice.call(forms).forEach(function (form) {
    form.addEventListener('submit', function (event) {
        if (!form.checkValidity()) {
            event.preventDefault()
            event.stopPropagation()
        }
        form.classList.add('was-validated')
    }, false)
});
*/