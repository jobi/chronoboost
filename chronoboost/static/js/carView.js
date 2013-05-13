function CarView() {
    this._init();
}

CarView.prototype = {
    _init : function() {
        this._view = document.createElement("tr");

        this._positionView = document.createElement("td");
        this._view.appendChild(this._positionView);

        this._driverView = document.createElement("td");
        this._view.appendChild(this._driverView);

        this._gapView = document.createElement("td");
        this._view.appendChild(this._gapView);

        this._intervalView = document.createElement("td");
        this._view.appendChild(this._intervalView);

        this._lapTimeView = document.createElement("td");
        this._view.appendChild(this._lapTimeView);

        this._sector1View = document.createElement("td");
        this._view.appendChild(this._sector1View);

        this._sector2View = document.createElement("td");
        this._view.appendChild(this._sector2View);

        this._sector3View = document.createElement("td");
        this._view.appendChild(this._sector3View);
    },

    updateFromModel : function(carModel) {
        this._positionView.innerHTML = carModel.Position;
        this._driverView.innerHTML = carModel.Driver;
        this._gapView.innerHTML = carModel.Gap;
        this._intervalView.innerHTML = carModel.Interval;

        this._lapTimeView.innerHTML = carModel.LapTime;
        this._updateStatus(this._lapTimeView, carModel.LapTimeStatus);

        this._sector1View.innerHTML = carModel.Sector1;
        this._updateStatus(this._sector1View, carModel.Sector1Status);

        this._sector2View.innerHTML = carModel.Sector2;
        this._updateStatus(this._sector2View, carModel.Sector2Status);

        this._sector3View.innerHTML = carModel.Sector3;
        this._updateStatus(this._sector3View, carModel.Sector3Status);
    },

    _updateStatus : function(elt, timeStatus) {
        switch (timeStatus) {
            case 3:
                elt.setAttribute("class", "personal_best");
                break;
            case 4:
                elt.setAttribute("class", "record");
                break;
            default:
                elt.removeAttribute("class");
                break;
        }
    },

    get view() {
        return this._view;
    }
}
