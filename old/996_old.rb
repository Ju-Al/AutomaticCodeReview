<%= render 'shared/navbar_modal' %>
          <%= link_to 'Curriculum', courses_path(ref: 'homenav'), class: 'nav-link' %>

<div class="navbar navbar-expand-md navbar-light">
  <div class="navbar-brand">
    <%= render 'shared/logo' %>
  </div>

  <button class="navbar-toggler navbar-toggler-right collapsed" type="button" data-toggle="modal" data-target="#navbar-supported-content" aria-controls="navbar-supported-content" aria-expanded="false" aria-label="toggle-navigation">
    <span class="navbar-toggler-icon open"></span>
    <span class="sr-only">Open navigation</span>
  </button>

  <% if user_signed_in? %>
    <nav class="collapse navbar-collapse">
      <!-- Full Width Navbar -->
      <ul class="navbar-nav ml-auto hidden-md-down">
        <li class="nav-item <%= active_class(courses_path) %>">
          <%= link_to 'My courses', track_path(current_user.track_id, ref: 'homenav'), class: 'nav-link' %>
        </li>
        <li class="nav-item dropdown js-dropdown">
          <a class="nav-link dropdown-toggle" href="#" id="navbarDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
            Community
          </a>
          <ul class="dropdown-menu dropdown-menu-right menu-options pb-0">
            <li>
              <%= link_to forum_link, class: 'drop-down-link accent', target: '_blank', rel: 'noreferrer' do %>
                <span class="drop-down-logo fa fa-users mt-2"></span>Forum
              <% end %>
            </li>
            <li>
              <%= link_to chat_link, class: 'drop-down-link accent', target: '_blank', rel: 'noreferrer' do %>
                <span class="drop-down-logo fa fa-comment mt-2"></span>Chat
              <% end %>
            </li>
          </ul>
        </li>

        <li class="divider nav-link">|</li>
        <%= render 'shared/user_dropdown' %>
      </ul>
  <% else %>
    <nav class="collapse navbar-collapse">
      <ul class="navbar-nav ml-auto">
        <li class="nav-item <%= active_class(courses_path) %>">
          <%= link_to 'Curriculum', track_path(1, ref: 'homenav'), class: 'nav-link' %>
        </li>
        <li class="nav-item dropdown js-dropdown">
          <a class="nav-link dropdown-toggle" href="#" id="navbarDropdown" role="button" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
            Community
          </a>
          <ul class="dropdown-menu dropdown-menu-right menu-options pb-0">
            <li>
              <%= link_to forum_link, class: "drop-down-link accent", target: '_blank', rel: 'noreferrer' do %>
                <span class="drop-down-logo fa fa-users mt-2"></span>Forum
              <% end %>
            </li>
            <li>
              <%= link_to chat_link, class: 'drop-down-link accent', target: '_blank', rel: 'noreferrer' do %>
                <span class="drop-down-logo fa fa-comment mt-2"></span>Chat
              <% end %>
            </li>
          </ul>
        </li>
        <li class="nav-item <%= active_class(about_path) %>">
          <%= link_to 'About', about_path(ref: 'homenav'), class: 'nav-link' %>
        </li>
        <li class="nav-item <%= active_class(faq_path) %>">
          <%= link_to "FAQ", faq_path(ref: 'homehav'), class: 'nav-link' %>
        </li>

        <li class="divider nav-link">|</li>

        <li class="nav-item <%= active_class(sign_up_path) %>">
          <%= link_to 'Sign Up', sign_up_path(ref: 'homenav'), class: 'nav-link' %>
        </li>
        <li class="nav-item <%= active_class(login_path) %>">
          <%= link_to 'Log In', login_path(ref: 'homenav'), class: 'nav-link' %>
        </li>
      </ul>
    </nav>
  <% end %>
    </ul>
  </nav>
</div>
