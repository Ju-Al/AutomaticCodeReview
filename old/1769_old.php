    var siteadminemail = $('#siteadminemail').val();
    var siteadminname = $('#siteadminname').val();
    var emailmethod = $('#emailmethod input:checked').val();
    var emailsmtphost = $('#emailsmtphost').val();
    var emailsmtpuser = $('#emailsmtpuser').val();
    var emailsmtppassword = $('#emailsmtppassword').val();
    var emailsmtpssl = $('#emailsmtpssl input:checked').val();
    var emailsmtpdebug = $('#emailsmtpdebug input:checked').val();

    if (siteadminemail != '<?= $siteadminemail ?>' ||
        siteadminname != '<?= $siteadminname ?>' ||
        emailmethod != '<?= $emailmethod ?>' ||
        emailsmtphost != '<?= $emailsmtphost ?>' ||
        emailsmtpuser != '<?= $emailsmtpuser ?>' ||
        emailsmtppassword != '<?= $emailsmtppassword ?>' ||
        emailsmtpssl != '<?= $emailsmtpssl ?>' ||
        emailsmtpdebug != '<?= $emailsmtpdebug ?>') {
        $('#settingschangedwarning').show();
    }
</script>