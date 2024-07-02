<section class="content-header">
    <div class="row mb-2">
      <div class="col-sm-6">
        <% content_for :title, "New Individuals Request - #{current_partner.name}" %>
        <h1><i class="fa fa-users"></i>&nbsp;&nbsp;
          New Request
          <small>for <%= current_partner.name %></small>
        </h1>
      </div>
      <div class="col-sm-6">
        <ol class="breadcrumb float-sm-right">
          <li class="breadcrumb-item"><a href="<%= partner_user_root_path %>"><i class="fa fa-home fa-lg"></i></a></li>
          <li class="breadcrumb-item"><a href="#">New Invididuals Order Request</a></li>
        </ol>
      </div>
    </div>
  </div><!-- /.container-fluid -->
</section>

<section class="content">
  <div class="container-fluid">
    <div class="row">
      <div class="col-md-12">
        <!-- Default box -->
        <div class="card">
          <div class="card-body">

            <%= @errors&.full_messages %>

            <%= simple_form_for @request, url: partners_individuals_requests_path, html: {role: 'form', class: 'form-horizontal'} do |form| %>

              <%= form.input :comments, label: "Comments:", as: :text, class: "form-control", wrapper: :input_group %>

              <table class='table'>
                <thead>
                  <tr>
                    <th>Item Requested</th>
                    <th>Number of Individuals</th>
                    <th></th>
                  </tr>
                </thead>

                <tbody class='fields'>
                  <%= render 'partners/individuals_requests/item_request', form: form %>
                </tbody>
              </table>

              <div>
                <%= add_item_button('Add Another Item') do %>
                  <%= render 'partners/individuals_requests/item_request', form: form %>
                <% end %>
              </div>

              <hr>

              <div class="card-footer">
                <!-- TODO(chaserx): we should add some js to prevent submission if the items selected are the blank option or any item has an empty quantity -->
                <%= form.submit("Submit Essentials Request", class: "btn btn-primary") %>
                <%= link_to "Cancel Request", partners_requests_path, class: "btn btn-danger" %>
              </div>
            <% end %>
        </div>
      </div>
    </div>
  </div>
</section>

<script type="text/javascript">
  $(document).ready(function() {
    $(document).on('click', '[data-add-target][data-add-template]', (event) => {
      var button = $(event.target)
      var target = button.data('add-target');
      var template = button.data("add-template");
      var templateId = new Date().getTime();
      var rendered = template.replace(/([\[_])([0-9]+)([\]_])/g, "$1" + templateId + '$3');

      $(target).append(rendered);

      event.preventDefault();
    });

    $(document).on('click', '[data-remove-item]', (event) => {
      var button = $(event.target)
      var wrapper = button.closest('tr')
      if (button.data("remove-item") === "soft") {
        wrapper.hide();
        button.prev('input[type=hidden]').val('1');
      } else {
        wrapper.remove()
      }
      event.preventDefault();
    });
  })
</script>
