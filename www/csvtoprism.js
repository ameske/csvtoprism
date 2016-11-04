sampleGroups = [];
experiment = {};
adjustedExperiment = [];

function resetInterface() {
  content = document.getElementById("content");
  if (content.hasChildNodes()) {
    while(content.childNodes.length > 1) {
      content.removeChild(content.lastChild);
    }
  }

  sampleGroups = [];
  experiment = {};
  adjustedExperiment = [];
}

function submitXLSFile() {
  var r = new XMLHttpRequest();

  var file = document.getElementById("xlsfile").files[0];
  
  var formData = new FormData();
  formData.append("xlsfile", file);

  r.open("POST", "http://localhost:41586/upload", true);
  r.onreadystatechange = function() {
      if (r.readyState == XMLHttpRequest.DONE) {
        if (r.status != 200) {
          alert("Something Went Wrong: " + r.responseText);
          return;
        }
        var response = JSON.parse(r.responseText);
        createAssociationMenu(response);
      }
  };
  r.send(formData);
}

function createAssociationMenu(data) {
  resetInterface();

  experiment = data;

  var sampleSelector = document.createElement("SELECT");
  sampleSelector.setAttribute("id", "availableSamples");

  for (let sample of data.experiment) {
    var o = document.createElement("OPTION");
    o.value = sample.name;
    o.text = sample.name;
    sampleSelector.add(o);
  }

  var addSampleGroupButton = document.createElement("BUTTON");
  addSampleGroupButton.setAttribute("id", "addNewSampleGroup");
  addSampleGroupButton.setAttribute("type", "button");
  addSampleGroupButton.innerHTML = "Create New Sample Group";
  addSampleGroupButton.addEventListener("click", createNewSampleGroup);

  var submitExperimentButton = document.createElement("BUTTON");
  submitExperimentButton.setAttribute("id", "submitExperiment");
  submitExperimentButton.setAttribute("type", "button");
  submitExperimentButton.innerHTML = "Submit Experiment";
  submitExperimentButton.addEventListener("click", submitExperiment);

  var content = document.getElementById("content");
  content.appendChild(sampleSelector);
  content.appendChild(addSampleGroupButton);
  content.appendChild(submitExperimentButton);
}

function createNewSampleGroup() {
    var s = document.getElementById("availableSamples");

    var control = s.options[s.selectedIndex];

    // Remove the newly selected control from all sample groups;
    s.remove(control);
    for (let sg of sampleGroups) {
      removeSampleByName(sg, control.value);
    }

    // Create the HTML Objects representing the new sample group
    createNewSampleGroupHTML(control.value, s.options);
}

function createNewSampleGroupHTML(control, remainingSamples) {
    var d = document.createElement("DIV");
    d.setAttribute("id", control);

    var l = document.createElement("LABEL");
    l.innerHTML = control + ":";

    // Create this sample groups select options
    var sampleGroupSelect = document.createElement("SELECT");
    sampleGroupSelect.setAttribute("id", control + "_select");
    for(i = 0; i < remainingSamples.length; i++) {
      var o = document.createElement("OPTION");
      o.value = remainingSamples[i].value;
      o.text = remainingSamples[i].text;
      sampleGroupSelect.add(o);
    }

    // Add the new sample group to our in memory representation
    sampleGroups.push(sampleGroupSelect);

    var addSample = document.createElement("BUTTON");
    addSample.innerHTML = "Add Sample To Group";
    addSample.addEventListener("click", function(){
      // Get the select box containing the sample we wish to add
      var s = document.getElementById(control+"_select");
      sample = s.options[s.selectedIndex].value;

      // Find the list for this sample group
      var list = document.getElementById(control+"_sampleList");

      // Create the list item and add it to the list
      var li = document.createElement("LI");
      li.innerHTML = sample;
      list.appendChild(li)

      // Also add it to our in memory representation
      for(i = 0; i < adjustedExperiment.length; i++) {
	 if (adjustedExperiment[i].Control = control) {
		adjustedExperiment[i].Samples.push(sample);
	 }
      }

      // Remove this option from all other sample groups
      var s = document.getElementById("availableSamples");
      removeSampleByName(s, sample);
      for (let sg of sampleGroups) {
        removeSampleByName(sg, sample);
      }
    });

    var deleteSampleGroup = document.createElement("BUTTON");
    deleteSampleGroup.innerHTML = "Delete Sample Group";
    deleteSampleGroup.addEventListener("click", function(){
	  // Remove the sample group from our in memory representation of the experiment
	  var removeIndex = -1;
	  for(i = 0; i < adjustedExperiment.length; i++) {
		if (adjustedExperiment[i].Control = control) {
			removeIndex = -1;
			break;
		}
	  }
	  adjustedExperiment.splice(removeIndex, 1); 
	  

          // Find all the samples added to this group so we can add them back
          var list = document.getElementById(control+"_sampleList");
          for (i = 0; i < list.childNodes.length; i++) {
            var sample = list.childNodes[i].innerHTML;
            addSampleByName(document.getElementById("availableSamples"), sample);
            for (let s of sampleGroups) {
              addSampleByName(s, sample);
            }
          }

          // Add the control sample back to everything
          addSampleByName(document.getElementById("availableSamples"), control);
          for (let s of sampleGroups) {
            addSampleByName(s, control);
          }
          var element = document.getElementById(control);
          element.parentNode.removeChild(element);
    });

    // Create the list representing what we have in the sample group
    var sampleList = document.createElement("UL");
    sampleList.setAttribute("id", control + "_sampleList");

    // Build up the div
    l.appendChild(sampleGroupSelect);
    l.appendChild(addSample);
    l.appendChild(deleteSampleGroup);
    d.appendChild(l);
    d.appendChild(document.createElement("BR"));
    d.appendChild(sampleList);

    // Add the sample group HTML to the DOM
    var content = document.getElementById("content");
    content.appendChild(d);

    // And add it to our in memory representation as well
    adjustedExperiment.push({"Control": control, "Samples": []});
}

// removeSampleByName removes the sample with the given name from the select options list
// 
// Parameters:
//      s - select DOM object
//      name - string
function removeSampleByName(s, name) {
  var options = s.options;

  for (i = 0; i < options.length; i++) {
    if (options[i].value == name) {
      s.remove(i);
      return;
    }
  }
}

// addSampleByName adds a sample as an option to the given select options list
// 
// Parameters:
//      s - select DOM object
//      name - string
function addSampleByName(s, name) {
  var o = document.createElement("OPTION");
  o.value = name;
  o.text = name;

  s.options.add(o);
}


function submitExperiment() {
  // Ensure that all samples have been allocated
  if (document.getElementById("availableSamples").options.length != 0) {
	alert("You have to assign all samples first before submitting the experiment");
	return;
  }

  console.log("Preparing to Submit Experiment: " + experiment.name);
  console.log(experiment);
  console.log(adjustedExperiment);

  var submitData = {"name": experiment.name, "samples": []};

  for (let sg of adjustedExperiment) {
    var sgData = {
      "control": {},
      "experimental": []
    };
    
    sgData.control = findSample(experiment.experiment, sg.Control);

    for (let sample of sg.Samples) {
      sgData.experimental.push(findSample(experiment.experiment, sample));
    }

    submitData.samples.push(sgData);
  }

  console.log(submitData);

  var r = new XMLHttpRequest();

  r.open("POST", "http://localhost:41586/generate", true);
  r.setRequestHeader("Content-Type", "application/json");
  r.onreadystatechange = function() {
      if (r.readyState == XMLHttpRequest.DONE) {
        if (r.status != 200) {
          alert("Something Went Wrong: " + r.responseText);
          return;
        }
        alert(r.responseText);
      }
  };
  r.send(JSON.stringify(submitData));
}

// findSample looks through a list of JSON objects representing raw samples
// 
// The JSON looks like:
//
//  {
//    name: "SampleName",
//    data: [0, 1, 2]
//  }
function findSample(samples, name) {
  for (let s of samples) {
    if (s.name == name) {
      return s;
    }
  }

  return null;
}
