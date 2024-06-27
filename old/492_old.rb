<div class="col-md-12 col-sm-6 col-xs-12" style="padding: 0">
  <%= presenter.gravatar_image(size: 200,
                               class: "img-rounded media-object pull-left") %>
</div>

<div class="col-md-12 col-sm-6 col-xs-12 user-summary" style="padding: 0">
  <ul>
    <h3><%= presenter.display_name %></h3>
    <% if presenter.has_title? %>
      <li><%= presenter.title_list %></li>
    <% end %>
    <% if presenter.country? %>
      <li><i class="fa fa-globe fa-lg"></i> <%= presenter.country %></li>
    <% end %>
    <% if presenter.latitude? && presenter.longitude? %>
      <li><i class="fa fa-clock-o fa-lg"></i> <%= NearestTimeZone.to(presenter.latitude, presenter.longitude) %></li>
    <% end %>

    <li>
      <i class="fa fa-calendar-o fa-lg"></i>
      Member for <%= presenter.object_age_in_words %>
    </li>

    <% if presenter.github_profile_url? %>
      <li>
        <i class="fa fa-github-alt fa-lg"></i>
        <%= presenter.github_link %>
      </li>
    <% end %>

    <% if presenter.display_email %>
      <li><i class="fa fa-envelope fa-lg"></i> <%= presenter.email_link %></li>
    <% end %>

    <% if presenter.display_hire_me %>
      <li>
        <%= link_to 'Hire me', user_path(presenter.user),
          { remote: true, class: 'user-profile-btn',
            data: { toggle: 'modal', target: '#modal-window'} } %>
      </li>
    <% end %>

    <% if presenter.object.eql?(current_user) %>
      <li>
        <%= link_to 'Edit', edit_user_registration_path,
          class: 'user-profile-btn',
          type: 'button' %>
      </li>
    <% end %>
  </ul>
</div>
