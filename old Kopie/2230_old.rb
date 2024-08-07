<section class="content-header">
        <% if @partner.notes.present? %>
          <%= render "notes", partner: @partner %>
        <% end %>
  <div class="container-fluid">
    <div class="row mb-2">
      <div class="col-sm-6">
        <% content_for :title, "Partners - #{@partner.name} - Agencies - #{current_organization.name}" %>
        <h1>
          Partner info
          <small>for <%= @partner.name %></small>
        </h1>
      </div>
      <div class="col-sm-6">
        <ol class="breadcrumb float-sm-right">
          <li class="breadcrumb-item"><%= link_to(dashboard_path) do %>
              <i class="fa fa-dashboard"></i> Home
            <% end %>
          </li>
          <li class="breadcrumb-item"><%= link_to "Partners", (partners_path) %></li>
          <li class="breadcrumb-item"><a href="#"><%= @partner.name %></a></li>
        </ol>
      </div>
    </div>
  </div><!-- /.container-fluid -->
</section>

<section class="content">
  <div class="container-fluid">
    <div class="row">
      <div class="col-12">
        <!-- Default box -->
        <% if @partner.uninvited? %>
          <%= render "uninvited_header", partner: @partner %>
        <% else %>
          <%= render "show_header", partner: @partner, impact_metrics: @impact_metrics, partner_distributions: @partner_distributions %>
        <% end %>

        <section class="card card-info card-outline">
          <div class="card-header">
            <h2 class="card-title">Contact Information</h2>
          </div>
          <div class="card-body p-3">
            <div class="row px-2">
              <div class="col-lg-4 col-sm-12">
                <div>
                  <h4> Executive Director </h4>
                  <div>
                    <label>Name:</label> <%= optional_data_text(@partner.profile.executive_director_name) %>
                  </div>
                  <div>
                    <label>Email:</label> <%= optional_data_text(@partner.profile.executive_director_email) %>
                  </div>
                  <div>
                    <label>Phone:</label> <%= optional_data_text(@partner.profile.executive_director_phone) %>
                  </div>
                </div>
              </div>
              <div class="col-lg-4 col-sm-12">
                <div>
                  <h4> Program Manager </h4>
                  <div>
                    <label>Name:</label> <%= optional_data_text(@partner.profile.program_contact_name) %>
                  </div>
                  <div>
                    <label>Email:</label> <%= optional_data_text(@partner.profile.program_contact_email) %>
                  </div>
                  <div>
                    <label>Phone:</label> <%= optional_data_text(@partner.profile.program_contact_phone) %>
                  </div>
                  <div>
                    <label>Mobile:</label> <%= optional_data_text(@partner.profile.program_contact_mobile) %>
                  </div>
                </div>
              </div>
              <div class="col-lg-4 col-sm-12">
                <div>
                  <h4> Pickup Person </h4>
                  <div>
                    <label>Name:</label> <%= optional_data_text(@partner.profile.pick_up_name) %>
                  </div>
                  <div>
                    <label>Email:</label> <%= optional_data_text(@partner.profile.pick_up_email) %>
                  </div>
                  <div>
                    <label>Phone:</label> <%= optional_data_text(@partner.profile.pick_up_phone) %>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="card card-info card-outline">
          <div class="card-header">
            <h2 class="card-title">Notes</h2>
          </div>
          <div class="card-body p-2">
            <div class="row">
              <div class="col-xs-12 col-sm-12" id="partner-notes">
                <% if @partner.notes %>
                  <p><%= simple_format(@partner.notes) %></p>
                <% else %>
                  <p> None provided </p>
                <% end %>
              </div>
            </div>
          </div>
        </section>

        <% if @partner.documents.present? %>
          <%= render "documents", partner: @partner %>
        <% end %>
      </div>
    </div>
  </div>
</section>

<section class="content">
  <div class="container-fluid">
    <div class="row">
      <div class="col-12">
        <!-- Default box -->
        <div class="card card-primary card-outline">
          <div class="card-header">
            <h2 class="card-title">Prior Distributions</h2>
          </div>
          <div class="card-body p-0">
            <div class="tab-content" id="custom-tabs-three-tabContent">
              <table class="table">
                <thead>
                <tr>
                  <th>Date</th>
                  <th>Source Inventory</th>
                  <th>Total items</th>
                  <th class="text-right">Actions</th>
                </tr>
                </thead>
                <tbody>
                <% @partner_distributions.each do |dist| %>
                  <tr>
                    <td><%= dist.created_at.strftime("%m/%d/%Y") %></td>
                    <td><%= dist.storage_location.name %></td>
                    <td><%= dist.line_items.total %></td>
                    <td class="text-right">
                      <%= view_button_to distribution_path(dist) %>
                      <%= print_button_to print_distribution_path(dist, format: :pdf) %>
                    </td>
                  </tr>
                  </tbody>
                <% end %>
                </table>
            </div><!-- /.box-body.table-responsive -->
          </div>
        </div>
      </div>
    </div>
  </div>

  <div id="addUserModal" class="modal fade">
    <div class="modal-dialog">
      <!-- Modal content-->
      <div class="modal-content">
        <div class="modal-header">
          <h4 class="modal-title">Add a User or Send a Reminder for <%= @partner.name %></h4>
        </div><!-- modal-header -->
        <div class="modal-body">
          <div class="box-body">
            <p>
              This will invite an additional member of the partner organization to Partnerbase unless a user with that
              exists in the Partnerbase system. If a user with that email exists it will send them a password reminder.
            </p>
            <br/>
            <%= form_tag re_invite_partner_path do %>
              <div class="input-group">
                <span class="input-group-text" id="spn_env_fa_icon"><%= fa_icon "envelope" %></span>
                <input type="email" name="email" class="form-control" placeholder="Email" aria-describedby="spn_env_fa_icon">
                <%= hidden_field_tag :partner, @partner.id %><br>
              </div>
              <br>
              <%= submit_button({text: "Invite User", icon: "envelope"}) %>
            <% end # form %>
          </div><!-- box-body -->
        </div><!-- modal-body -->
      </div><!-- modal-content -->
    </div><!-- modal-dialog -->
  </div><!-- addUserModal -->
</section>
