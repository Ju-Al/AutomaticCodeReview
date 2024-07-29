<div class="container">
  <div class="row">
    <div class="col-md-10 offset-md-1 col-lg-8 offset-lg-2">
      <div class="card-main profile-card">
        <div class="row">
          <div class="col-md-5 align-items-center text-center">
            <div class="user-avatar">
              <%= image_tag gravatar_url(user, 80), alt: 'gravatar' %>
            </div>

            <p class="welcome-text"> Welcome back <%=  user.username %>! </p>
          </div>

          <div class="col-sm-1"></div>

          <div class="col-md-6 info">
            <div class="learning-goal-wrapper">
              <p class="camel bold learning-goal-title">Learning Goal</p>
              <p class="learning-goal">
                <%= user.learning_goal || ''  %>

                <% if user.learning_goal == nil %>
                <%= link_to "Settings", edit_user_registration_path, class: 'accent' %>
                <% end %>



              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
