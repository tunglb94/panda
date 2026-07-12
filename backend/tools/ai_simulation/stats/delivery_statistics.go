package stats

import "github.com/fairride/ai_simulation/domain/entity"

// DeliveryStatistics covers only trips with Kind == KindDelivery — the
// PHẦN 1 delivery-lifecycle counterpart to dispatch_statistics.json's
// ride-and-delivery-combined dispatch view.
type DeliveryStatistics struct {
	Requested            int     `json:"requested"`
	Accepted             int     `json:"accepted"`
	Rejected             int     `json:"rejected"`
	Cancelled            int     `json:"cancelled"`
	Completed            int     `json:"completed"`
	AveragePickupMinutes float64 `json:"average_pickup_minutes"`
	AverageDeliveryMinutes float64 `json:"average_delivery_minutes"` // pickup -> delivered transit time
	AverageDistanceKM    float64 `json:"average_distance_km"`
	AverageWeightKg      float64 `json:"average_weight_kg"`
}

func (c *Collector) BuildDeliveryStatistics(trips []*entity.SimTrip) DeliveryStatistics {
	var out DeliveryStatistics
	var pickupSum, transitSum, distanceSum, weightSum float64
	var pickupN, transitN int

	for _, t := range trips {
		if t.Kind != entity.KindDelivery {
			continue
		}
		out.Requested++
		distanceSum += t.DistanceKM
		weightSum += t.PackageWeightKg
		switch t.Outcome {
		case entity.OutcomeCompleted:
			out.Completed++
			out.Accepted++
			if t.PickupMinutes > 0 {
				pickupSum += t.PickupMinutes
				pickupN++
			}
			if t.DeliveryTransitMinutes > 0 {
				transitSum += t.DeliveryTransitMinutes
				transitN++
			}
		case entity.OutcomeRejected:
			out.Rejected++
		case entity.OutcomeCancelled:
			out.Cancelled++
		}
	}

	out.AveragePickupMinutes = avg(pickupSum, pickupN)
	out.AverageDeliveryMinutes = avg(transitSum, transitN)
	out.AverageDistanceKM = avg(distanceSum, out.Requested)
	out.AverageWeightKg = avg(weightSum, out.Requested)
	return out
}
