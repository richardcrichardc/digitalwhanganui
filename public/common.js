
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
});
