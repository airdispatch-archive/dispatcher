$(document).ready(function() {
	var counter = 0;
	var advanced = false;

	function createRemoveEvents() {
		$(".btn-remove").click(function() {
			$(this).parent().remove();
			return false;
		});
		counter = $(".content-control").size();
	}
	createRemoveEvents();

	$(".btn-add").click(function() {
		addType("airdispat.ch/generic", false)
		return false;
	});

	$(".btn-menu a").click(function() {
		addType($(this).attr("data-add"), 
			$(this).attr("data-disabled")===undefined);
		console.log($(this).attr("data-disabled"));
	});

	function addType(theType, disabled) {
		if(theType != undefined) {
			if (theType.indexOf("/") >= 0) {
				$(".all-controls").append(createControl(theType, disabled));
				createRemoveEvents();
			} else {
				$("[data-add~='" + theType + "']").parent().siblings().children("a").each(function() {
					addType($(this).attr("data-add"), disabled);
				});
			}
		}
	}

	function createControl(theType, disabled) {
		var output = "<div class='controls controls-row content-control'>";
		var input1 = "<input type='text' id='" + theType + "' class='span3 control-name' name='content[" + counter + "][0]' value='" + theType + "'";
		input1 += (disabled?"readonly='readonly'>":"data-advanced='true' >");

		var input2 = "";
		if (theType.indexOf("content") >= 0) { 
			input2 = "<textarea style='height: 200px' id='" + theType + "' class='span5' name='content[" + counter + "][1]'></textarea>";
		} else {
			input2 = "<input type='text' id='" + theType + "' class='span5' name='content[" + counter + "][1]'>";
		}

		var removeButton = "<button class='btn btn-danger span1 btn-remove'><i class='icon-remove'></i></button>";
		var close = "</div>";
		return (output + input1 + input2 + removeButton + close);
	}

	$(".btn-toggle").click(function() {
		advanced = !advanced;
		if(advanced) {
			$(".toggle-icon").removeClass("icon-unchecked").addClass("icon-check")
			$(".control-name[data-advanced!='true']").removeAttr("readonly")
		} else {
			$(".toggle-icon").removeClass("icon-check").addClass("icon-unchecked")
			$(".control-name[data-advanced!='true']").attr("readonly", true)
		}
		return false
	});
})