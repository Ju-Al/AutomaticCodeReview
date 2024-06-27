<!DOCTYPE html>
  <!-- iCheck -->
  .checkbox label{
    padding-left: 0px;
  }
</style>
<body id="devise" class="hold-transition login-page">
<div class="login-box">
  <div class="login-logo">

    <img id="MastheadLogo" src="/img/diaper-base-logo-full-white-text.svg" style="width: 80%; height: 80%; " alt="Diaperbase" title="Diaperbase" class="serv_icon"/>
  </div>
  <!-- /.login-logo -->
  <div class="login-box-body">
    <div class="form-group has-feedback">
      <%= yield %>
  <!-- /.login-box-body -->
</div>
<!-- /.login-box -->
<script src="/assets/icheck.min.js"></script>
<script>
  $(function () {
    $('input').iCheck({
      checkboxClass: 'icheckbox_square-blue',
      radioClass: 'iradio_square-blue',
      increaseArea: '20%' // optional
    });
  });
</script>
<html>
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Diaperbase | Log in</title>
  <!-- Tell the browser to be responsive to screen width -->
  <meta content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" name="viewport">
  <!-- Bootstrap 3.3.7 -->
  <%= csrf_meta_tags %>
  <%= javascript_include_tag 'application' %>
  <%= stylesheet_link_tag    'application', media: 'all' %>
  <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Source+Sans+Pro:300,400,600,700,300italic,400italic,600italic">
  <style>
    .checkbox label{
      padding-left: 0px;
    }
  </style>
  <script>
    $(document).ready(function() {
      <% if Rails.env.staging? %>
        // Prevents users from closing the modal by clicking outside of it or pressing the esc key
        $('#warningModal').modal({
          backdrop: 'static',
          keyboard: false
        });
        // Adds click event handler on the checkbox
        $('#defaultCheck1').click(function(){
          // If the checkbox is checked, enable the continue button
          if($(this).is(':checked')){
            $('.continue-btn').attr("disabled", false);
          } else{
            $('.continue-btn').attr("disabled", true);
          }
        });
      <% end %>
    });
  </script>
</head>
<body id="devise" class="hold-transition login-page <%= Rails.env.staging? ? 'login-page-test' : '' %>">
  <div class="login-box">
    <div class="login-logo">
      <img id="MastheadLogo" src="/img/diaper-base-logo-full-white-text.svg" style="width: 80%; height: 80%; " alt="Diaperbase" title="Diaperbase" class="serv_icon"/>
    </div>
    <!-- /.login-logo -->
    <div class="login-box-body">
      <div class="form-group has-feedback">
        <%= yield %>
      </div>
    </div>
    <!-- /.login-box-body -->
  </div>
  <!-- /.login-box -->

  <!-- Modal -->
  <div class="modal fade" id="warningModal" tabindex="-1" role="dialog" aria-labelledby="warningModalLabel" aria-hidden="true">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h3 class="modal-title" id="warningModalLabel"><b>This site is for TEST purposes only!</b></h3>
        </div>
        <div class="modal-body">
          You're visiting diaperbase.org, a demo/test site for the full site at <a href="https://diaper.app">diaper.app</a>.<br>
          It is not safe to upload, enter or save any sensitive data here.<br>
          <div class="modal-body-warning-text">
            If you meant to login to your live account, go to <a href="https://diaper.app/users/sign_in">diaper.app</a>
          </div>
          <br>
          <div class="form-check">
            <input class="form-check-input" type="checkbox" value="" id="defaultCheck1">
            <label class="form-check-label" for="defaultCheck1">
              I Understand
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-warning continue-btn" data-dismiss="modal" disabled>Continue in Test Site</button>
        </div>
      </div>
    </div>
  </div>
</body>
</html>
