<div class="skills-header">
  <h2 class="skills-header__title">Skills Progress</h2>
  <p class="skills-header__content">
    If you're new to web development, take the following courses in the order. Otherwise feel free to start wherever interests you the most.
  </p>
</div>

<% courses.each do |course| %>
  <%= render 'skill', course: course %>
<% end %>

<p class="text-center">You are currently enrolled in the <%= @track.title %> track. To view all available tracks <%= link_to 'Click here', tracks_path %></p>
