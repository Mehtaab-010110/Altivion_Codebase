from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime, timezone

class Signal(BaseModel):
    ts: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    sn: str = Field(alias="SN")
    uasid: Optional[str] = Field(default=None, alias="UASID")
    drone_type: Optional[str] = Field(default=None, alias="DroneType")
    direction_deg: Optional[int] = Field(default=None, alias="Direction")
    speed_h_mps: Optional[float] = Field(default=None, alias="SpeedHorizontal")
    speed_v_mps: Optional[float] = Field(default=None, alias="SpeedVertical")
    lat: Optional[float] = Field(default=None, alias="Latitude")
    lon: Optional[float] = Field(default=None, alias="Longitude")
    height_m: Optional[float] = Field(default=None, alias="Height")
    operator_lat: Optional[float] = Field(default=None, alias="OperatorLatitude")
    operator_lon: Optional[float] = Field(default=None, alias="OperatorLongitude")

    class Config:
        populate_by_name = True  # allow snake_case or aliased names
