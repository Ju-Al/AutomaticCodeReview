<%= provide :title, 'Welcome - The resource for Chef cookbooks' %>
          <a href="https://wiki.opscode.com/display/ChefCN/Creating+New+Cookbooks" target="_blank">Share your cookbooks</a>
<%= provide :description, "Supermarket is Chef's open-source community platform. Find, explore and view Chef cookbooks for all of your ops needs." %>

<div class="page withspace">
  <% if flash[:signout] %>
    <div class="signout-message">
      <p>You have been signed out of Supermarket but may still be signed in to your Chef account. Visit <%= link_to nil, chef_identity_url %> to sign out completely.</p>
    </div>
  <% end %>

  <section class="sign_up_in">
    <h2 class="text-center">Welcome to Supermarket. Find, explore and view Chef cookbooks for all of your ops needs.</h2>

    <p class="text-center">
      <%= link_to 'Sign Up For a Chef Account', chef_sign_up_url, class: 'button radius' %>
      <%= link_to 'Sign In With Your Chef Account', sign_in_path, class: 'button radius' %>
    </p>
  </section>

  <section class="welcome_resources">
    <div>
      <h2>Explore</h2>

      <ul>
        <li>
          <%= link_to "Browse Cookbooks", cookbooks_path %>
        </li>
        <li>
          <%= link_to 'Read the Chef Blog', chef_blog_url, target: '_blank' %>
        </li>
      </ul>
    </div>
    <div>
      <h2>Learn</h2>
      <ul>
        <li>
          <%= link_to 'Learn Chef', learn_chef_url, target: '_blank' %>
        </li>
        <li>
          <%= link_to 'Read the Chef Docs', chef_docs_url, target: '_blank' %>
        </li>
        <li>
          <%= link_to 'Community Guidelines', chef_docs_url('community_guidelines.html'), target: '_blank' %>
        </li>
        <li>
          <%= link_to 'How to Contribute', chef_docs_url('community_contributions.html'), target: '_blank' %>
        </li>
      </ul>
    </div>
    <div>
      <h2>Share</h2>
      <ul>
        <li>
          <a href="http://docs.chef.io/knife_cookbook_site.html#share" target="_blank">Share your cookbooks</a>
        </li>
        <li>
          <a href="irc://irc.freenode.net/chef">Chat on IRC at #chef on irc.freenode.net</a>
        </li>
        <li>
          <a href="https://groups.google.com/forum/#!forum/chef-supermarket" target="_blank">Join the Supermarket Mailing List</a>
        </li>
        <li>
          <a href="https://github.com/opscode/supermarket" target="_blank">Contribute to Supermarket</a>
        </li>
      </ul>
    </div>
  </section>

  <section class="cookbooks_stats">
    <h2>Community Stats</h3>
    <%= render 'resources/community_stats', cookbook_count: @cookbook_count, user_count: @user_count %>
  </section>
</div>
