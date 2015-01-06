
function unobfusicateEmailLinks() {
  $(".obf-email").each( function() {
    var $node = $(this);
    var obfEmail = $node.attr('data-obf-email');
    var obfEmailLabel = $node.attr('data-obf-email-label');
    var len = obfEmail.length;
    var label;
    var email = ""

    for (var i=0; i < len; i ++) {
      email += String.fromCharCode(obfEmail.charCodeAt(i) + 1);
    }

    if (obfEmailLabel) {
      label = obfEmailLabel;
    } else {
      label = email;
    }

    $node
      .text(label)
      .attr("href", "mailto:" + email);
  });
}

$( document ).ready(function() {
  setTimeout(unobfusicateEmailLinks, 25);

  // copy input between the various search boxes
  $(".search-input").keydown(function() {
    var that=this;
    // wait for the key to register in val()
    setTimeout(function() {
      var value = $(that).val();
      $(".search-input").each(function() {
        // only update if the value differs so as not to lose cursor position
        if ($(this).val() != value) {
          $(this).val(value);
        }
      })
    }, 1);
  });

});
