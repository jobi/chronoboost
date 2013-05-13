function WeatherView(rootElt) {
    this._init(rootElt);
}

WeatherView.prototype = {
    _init : function(rootElt) {
        var dl = document.createElement("dl");
        rootElt.appendChild(dl);
        dl.classList.add("dl-horizontal");

        var dt = document.createElement("dt");
        dt.innerHTML = "Air temperature";
        dl.appendChild(dt);

        this._airTemperaturView = document.createElement("dd");
        dl.appendChild(this._airTemperaturView);

        dt = document.createElement("dt");
        dt.innerHTML = "Track temperature";
        dl.appendChild(dt);

        this._trackTemperaturView = document.createElement("dd");
        dl.appendChild(this._trackTemperaturView);

        dt = document.createElement("dt");
        dt.innerHTML = "Wind speed";
        dl.appendChild(dt);

        this._windSpeedView = document.createElement("dd");
        dl.appendChild(this._windSpeedView);

        dt = document.createElement("dt");
        dt.innerHTML = "Wind direction";
        dl.appendChild(dt);

        this._windDirectionView = document.createElement("dd");
        dl.appendChild(this._windDirectionView);
    },

    updateFromWeather : function(weatherModel) {
        this._airTemperaturView.innerHTML = weatherModel.AirTemp + "C";
        this._trackTemperaturView.innerHTML = weatherModel.TrackTemp + "C";
        this._windSpeedView.innerHTML = weatherModel.WindSpeed + "m/s";
        this._windDirectionView.innerHTML = weatherModel.WindDirection;
    }
};
