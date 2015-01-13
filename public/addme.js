$( document ).ready(function() {

  /* Category Selection Dialog */

  // Show first major and minor category select
  $("#major-cat-select select").first().removeClass("hidden");
  $("#minor-cat-select select").first().removeClass("hidden");

  // Change visible major and minor category select when majormajor category is changed
  $("#select-major-major-cats").change(function (e){
    var majorMajorCat = $(e.target).prop("value");
    $("#major-cat-select select").addClass("hidden");
    $("#select-" + majorMajorCat + "-major-cats").removeClass("hidden")
        .val($("#select-" + majorMajorCat + "-major-cats option:first").val());
    $("#minor-cat-select select").addClass("hidden");
    $("#" + majorMajorCat + "-cat-select select:first").removeClass("hidden")
        .val($("#" + majorMajorCat + "-cat-select select:first option:first").val());
  });

  // Change visible minor category select when major category is changed
  $("#major-cat-select select").change(function (e){
    var majorMajorCat = $("#select-major-major-cats").val();
    var majorCat = $(e.target).val();
    $("#minor-cat-select select").addClass("hidden");
    $("#select-" + majorMajorCat + "-" + majorCat + "-minor-cats").removeClass("hidden");
  });

  // Add category when Add button in dialog is clicked
  $("#add-category").click(function() {
    $("#cat-modal").modal('hide');

    var majorMajorCat = $("#select-major-major-cats").val();
    var majorMajorCatName = $('#select-major-major-cats option[value="' + majorMajorCat + '"]').text();
    var majorCat = $("#select-" + majorMajorCat + "-major-cats").val();
    var majorCatName = $('#select-' + majorMajorCat + '-major-cats option[value="' + majorCat + '"]').text();

    var minorCat = $("#select-" + majorMajorCat + "-" + majorCat + "-minor-cats").val();
    var minorCatName = $('#select-' + majorMajorCat + "-" + majorCat + '-minor-cats option[value="' + minorCat + '"]').text();
    var cat = majorMajorCat + "." + majorCat + "." + minorCat;
    var catName = majorMajorCatName + " > " + majorCatName;
    if (minorCat != "none") {
        catName += " > " + minorCatName;
    }

    // Don't add a category more than once
    if ($("#" + majorMajorCat + "\\." + majorCat + "\\." + minorCat).length) {
      return;
    }

    var catEl = $('<span class="badge" id="' + cat + '">' + catName + ' <span class="glyphicon glyphicon-remove remove-category"></span></span>');

    $("#add-at-least-one").addClass("hidden");
    $("#selected-categories").append(catEl).removeClass("hidden");

    setAllCategories();
  });

  // Remove a previously added category when X on category badge clicked
  $( document ).on( "click", ".remove-category", function(e) {
    $(e.target).parent(".badge").remove();
    if ($("#selected-categories").children().length == 0) {
      $("#add-at-least-one").removeClass("hidden");
      $("#selected-categories").addClass("hidden");
    }
    setAllCategories();
  });

  function setAllCategories() {
    $("#categories").val($("#selected-categories").children().map(function() { return this.id; }).get().join());
  }


  // Auto generate Name in individual listings
  $("#adminFirstName, #adminLastName").keypress(function() {
    setTimeout(function() {
         $("#individual-name").text($("#adminFirstName").val() + ' ' + $("#adminLastName").val());
    }, 1);
  });

  // Hide or show Name input depending on listing type

  $("#isOrg").change(function() {
    var isOrg = $("#isOrg").val() == "1";
    $("#individual-name").toggleClass("hidden", isOrg);
    $("#organisation-name").toggleClass("hidden", !isOrg);
  });

  // file upload widget
  $('#uploadImage input').fileupload({
    url: '/uploadImage',
    formData: {},
    sequentialUploads: true,

    submit: function (e, data) {
      $('#uploadImage').addClass("disabled");
      $('#image').addClass("hidden");
      $('#noImage').addClass("hidden");
      $('#imageUploading').removeClass("hidden");
      //$('#cancelUpload').removeClass("hidden");
    },

    always: function (e, data) {
      var error;

      /*console.log(JSON.stringify(data.result, null, 4));

      var obj = data;
      for (var prop in obj) {
        console.log("o." + prop + " = " + obj[prop]);
      }*/


      if (data.jqXHR && data.jqXHR.status && data.jqXHR.status == 413) {
        error = "File too big - please use a file smaller than 10MB.";
      } else if (data.textStatus === "error") {
        error = "Upload failed - please try again.";
      } else if (data.result.Error) {
        error = data.result.Error;
      }

      if (error) {
        $('#image')
          .removeClass("hidden")
          .text(error);
        $('#imageInput').val('');
      } else {
        $('#image')
          .removeClass("hidden")
          .html('<img src="/image/' + data.result.Id + '/small">');
        $('#imageInput').val(data.result.Id);
        $('#removeImage').prop("disabled", false);
      }

      $('#imageUploading').addClass("hidden");
      $('#uploadImage').removeClass("disabled");
      //$('#cancelUpload').addClass("hidden");
    },




  });

  $("#removeImage").click(function() {
    $('#imageInput').val('');
    $('#image').addClass("hidden");
    $('#noImage').removeClass("hidden");
    $('#removeImage').prop("disabled", true);
  });


  setAllCategories();
});