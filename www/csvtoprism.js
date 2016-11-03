contentContainer = null;
sampleGroups = [];
experiment = {};

$(document).ready(function() {
  contentContainer = document.getElementById("content")
  $("#xlsfileform").submit(function(e) {
    e.preventDefault();

    var data = new FormData(this);

    $.ajax({
      url: "http://localhost:41586/upload",
      data: data,
      cache: false,
      contentType: false,
      processData: false,
      dataType: "json",
      type: "POST",
      success: function(data) {
        experiment = data;
        createAssociationMenu(experiment);
      }
    });

    return true;
  });
});

function createAssociationMenu(data) {
  resetInterface();

  var sampleSelector = document.createElement("SELECT");
  sampleSelector.setAttribute("id", "sampleGroupSelect");

  for (let sample of data.experiment) {
    var o = document.createElement("OPTION");
    o.value = sample.name;
    o.text = sample.name;
    sampleSelector.add(o);
  }

  var addSampleGroup = document.createElement("BUTTON");
  addSampleGroup.setAttribute("id", "addNewSampleGroup");
  addSampleGroup.setAttribute("type", "button");
  addSampleGroup.innerHTML = "Create New Sample Group";
  addSampleGroup.addEventListener("click", createNewSampleGroup);

  contentContainer.appendChild(sampleSelector);
  contentContainer.appendChild(addSampleGroup);
}

function createNewSampleGroup() {
    s = document.getElementById("sampleGroupSelect");

    control = s.options[s.selectedIndex];

    // Remove the newly selected control from all sample groups;
    s.remove(control);
    for (let s of sampleGroups) {
      removeSampleByName(s, control.value);
    }

    // Create the HTML Objects representing the new sample group
    createNewSampleGroupHTML(control.value, s.options);

}

function createNewSampleGroupHTML(control, remainingSamples) {
    d = document.createElement("DIV");
    d.setAttribute("id", control);

    l = document.createElement("LABEL");
    l.innerHTML = control + ":";


    // Create this sample groups select options
    sampleGroupSelect = document.createElement("SELECT");
    sampleGroupSelect.setAttribute("id", control + "_select");
    for(i = 0; i < remainingSamples.length; i++) {
      var o = document.createElement("OPTION");
      o.value = remainingSamples[i].value;
      o.text = remainingSamples[i].text;
      sampleGroupSelect.add(o);
    }

    // Add the new sample group to our in memory representation
    sampleGroups.push(sampleGroupSelect);

    addSample = document.createElement("BUTTON");
    addSample.innerHTML = "Add Sample To Group";
    addSample.addEventListener("click", function(){
      // Get the select box containing the sample we wish to add
      s = document.getElementById(control+"_select");
      sample = s.options[s.selectedIndex].value;

      // Find the list for this sample group
      list = document.getElementById(control+"_sampleList");

      // Create the list item and add it to the list
      li = document.createElement("LI");
      li.innerHTML = sample;
      list.appendChild(li)

      // Remove this option from all other sample groups
      s = document.getElementById("sampleGroupSelect");
      removeSampleByName(s, sample);

      for (let s of sampleGroups) {
        removeSampleByName(s, sample);
      }
    });

    deleteSampleGroup = document.createElement("BUTTON");
    deleteSampleGroup.innerHTML = "Delete Sample Group";
    deleteSampleGroup.addEventListener("click", function(){
          // Find all the samples added to this group so we can add them back
          list = document.getElementById(control+"_sampleList");
          for (i = 0; i < list.childNodes.length; i++) {
            sample = list.childNodes[i].innerHTML;
            addSampleByName(document.getElementById("sampleGroupSelect"), sample);
            for (let s of sampleGroups) {
              addSampleByName(s, sample);
            }
          }

          // Add the control sample back to everything
          addSampleByName(document.getElementById("sampleGroupSelect"), control);
          for (let s of sampleGroups) {
            addSampleByName(s, control);
          }
          var element = document.getElementById(control);
          element.parentNode.removeChild(element);
    });

    // Create the list representing what we have in the sample group
    sampleList = document.createElement("UL");
    sampleList.setAttribute("id", control + "_sampleList");

    // Build up the div
    l.appendChild(sampleGroupSelect);
    l.appendChild(addSample);
    l.appendChild(deleteSampleGroup);
    d.appendChild(l);
    d.appendChild(document.createElement("BR"));
    d.appendChild(sampleList);

    // Add the sample group HTML to the DOM
    contentContainer.appendChild(d);
}

function removeSampleByName(s, name) {
  options = s.options;

  for (i = 0; i < options.length; i++) {
    if (options[i].value == name) {
      s.remove(i);
      return;
    }
  }
}

function addSampleByName(s, name) {
  o = document.createElement("OPTION");
  o.value = name;
  o.text = name;

  s.options.add(o);
}

function resetInterface() {
  if (contentContainer.hasChildNodes()) {
    while(contentContainer.childNodes.length > 1) {
      contentContainer.removeChild(contentContainer.lastChild);
    }
  }

  sampleGroups = [];
}
