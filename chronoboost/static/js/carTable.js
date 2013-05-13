function CarTable(elt) {
    this._init(elt);
}

CarTable.prototype = {
    _init : function(elt) {
        this._rootElt = elt;

        this._cars = {};
    },

    _updateTable : function() {
        var carList = [];

        for (var key in this._cars) {
            carList.push(this._cars[key]);
        }

        carList.sort(function (carA, carB) {
                         return carA.model.Position - carB.model.Position;
                     });

        this._rootElt.innerHTML = "";

        console.log("Updating with carList " + carList.length);

        for (var i = 0; i < carList.length; ++i) {
            this._rootElt.appendChild(carList[i].view.view);
        }
    },

    updateFromCar : function(carModel) {
        var car;
        var updateTable = false;
        var key = carModel.Number;

        if (key in this._cars) {
            car = this._cars[key];
            updateTable = car.model.Position != carModel.Position;
            car.model = carModel;
        } else {
            car = { model: carModel,
                    view: new CarView() };
            this._cars[key] = car;
            updateTable = true;
        }

        car.view.updateFromModel(carModel);

        if (updateTable) {
            this._updateTable();
        }
    }
};
