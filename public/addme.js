$( document ).ready(function() {
  console.log("hello");

  /* Category Selection Dialog */

  // Show first minor category select
  $("#minor-cat-select select").first().removeClass("hidden");

  // Change visible minor category select when major category is changed
  $("#select-major-cats").change(function (e){
    var majorCat = $(e.target).prop("value");
    $("#minor-cat-select select").addClass("hidden");
    $("#select-" + majorCat + "-minor-cats").removeClass("hidden");
  });

  // Add category when Add button in dialog is clicked
  $("#add-category").click(function() {
    $("#cat-modal").modal('hide');

    var majorCat = $("#select-major-cats").prop("value");
    var majorCatName = $('#select-major-cats option[value="' + majorCat + '"]').text();
    var minorCat = $("#select-" + majorCat + "-minor-cats").prop("value");
    var minorCatName = $('#select-' + majorCat + '-minor-cats option[value="' + minorCat + '"]').text();
    var cat = majorCat + "." + minorCat;
    var catName = majorCatName + " / " + minorCatName;

    // Don't add a category more than once
    if ($("#" + majorCat + "\\." + minorCat).length) {
      return;
    }

    var catEl = $('<span class="badge" id="' + cat + '">' + catName + ' <span class="glyphicon glyphicon-remove remove-category"></span></span>');

    $("#add-at-least-one").addClass("hidden");
    $("#selected-categories").append(catEl);

    setAllCategories();
  });

  // Remove a previously added category when X on category badge clicked
  $( document ).on( "click", ".remove-category", function(e) {
    $(e.target).parent(".badge").remove();
    if ($("#selected-categories").length) {
      $("#add-at-least-one").removeClass("hidden");
    }
    setAllCategories();
  });

  function setAllCategories() {
    $("#categories").val($("#selected-categories").children().map(function() { return this.id; }).get().join());
  }

  $('#uploadImage input').fileupload({
    url: '/uploadImage',
    sequentialUploads: true,

    submit: function (e, data) {
      console.log('sub', data);
      $('#uploadImage').addClass("disabled");
      $('#image').addClass("hidden");
      $('#noImage').addClass("hidden");
      $('#imageUploading').removeClass("hidden");
      //$('#cancelUpload').removeClass("hidden");
    },

    done: function (e, data) {
      console.log('dn', data);
      $('#image')
        .removeClass("hidden")
        .text(data.result.Id);
      $('#imageUploading').addClass("hidden");
      $('#uploadImage').removeClass("disabled");
      //$('#cancelUpload').addClass("hidden");
      $('#removeImage').prop("disabled", false);
    }
  });

  $("#removeImage").click(function() {
    $('#image').addClass("hidden");
    $('#noImage').removeClass("hidden");
    $('#removeImage').prop("disabled", true);
  });

});